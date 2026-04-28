package crawler

import (
	"encoding/json"
	"fmt"
	"html"
	"regexp"
	"strings"

	"github.com/google/uuid"

	"troveler/db"
)

// SearchResponse represents the JSON response from the search API.
type SearchResponse struct {
	Found       int64        `json:"found"`
	OutOf       int64        `json:"out_of"`
	FacetCounts []FacetCount `json:"facet_counts"`
	Hits        []Hit        `json:"hits"`
}

// FacetCount represents a facet with its field name and values.
type FacetCount struct {
	FieldName string       `json:"field_name"`
	Counts    []FacetValue `json:"counts"`
}

// FacetValue represents a single facet value with its count.
type FacetValue struct {
	Value string `json:"value"`
	Count int64  `json:"count"`
}

// Hit represents a single search result hit.
type Hit struct {
	Document HitDocument `json:"document"`
}

// HitDocument contains the tool data from a search hit.
type HitDocument struct {
	Slug          string   `json:"slug"`
	Name          string   `json:"name"`
	Tagline       string   `json:"tagline"`
	Description   string   `json:"preview"`
	Language      string   `json:"language"`
	License       []string `json:"license"`
	ToolOfTheWeek bool     `json:"tool_of_the_week"`
}

// ParseSearchResponse parses the raw JSON search response.
func ParseSearchResponse(data []byte) (*SearchResponse, error) {
	var resp SearchResponse
	if err := json.Unmarshal(data, &resp); err != nil {
		return nil, fmt.Errorf("failed to parse search response: %w", err)
	}

	return &resp, nil
}

// DetailPage represents a parsed tool detail page.
type DetailPage struct {
	Tool          db.Tool
	Installations map[string]string
}

var (
	jsonLDRegex   = regexp.MustCompile(`<script type="application/ld\+json">(.*?)</script>`)
	installRegex  = regexp.MustCompile(`id="install" data-install="([\s\S]*?)"`)
	toolSlugRegex = regexp.MustCompile(`data-tool-slug="([^"]+)"`)
	titleRegex    = regexp.MustCompile(`<title>([^<]+)`)
	taglineRegex  = regexp.MustCompile(`id="tagline">([^<]+)`)
)

// JSONLD represents the structured data from a page's JSON-LD script.
type JSONLD struct {
	Context string           `json:"@context"`
	Graph   []map[string]any `json:"@graph"`
}

// ParseDetailPage parses a tool detail page from raw HTML.
func ParseDetailPage(data []byte) (*DetailPage, error) {
	page := &DetailPage{
		Installations: make(map[string]string),
	}

	jsonLDMatch := jsonLDRegex.FindStringSubmatch(string(data))
	if len(jsonLDMatch) < 2 {
		return nil, fmt.Errorf("no JSON-LD found")
	}

	var jsonLD JSONLD
	if err := json.Unmarshal([]byte(jsonLDMatch[1]), &jsonLD); err != nil {
		return nil, fmt.Errorf("failed to parse JSON-LD: %w", err)
	}

	var softwareApp map[string]any
	for _, item := range jsonLD.Graph {
		if item["@type"] == "SoftwareApplication" {
			softwareApp = item

			break
		}
	}

	if softwareApp == nil {
		return nil, fmt.Errorf("no SoftwareApplication in JSON-LD")
	}

	page.Tool.ID = uuid.New().String()

	if slugMatch := toolSlugRegex.FindStringSubmatch(string(data)); len(slugMatch) >= 2 {
		page.Tool.Slug = slugMatch[1]
	} else if id, ok := softwareApp["@id"].(string); ok {
		parts := strings.Split(strings.TrimSuffix(id, "/"), "/")
		page.Tool.Slug = parts[len(parts)-1]
	}

	if name, ok := softwareApp["name"].(string); ok {
		page.Tool.Name = name
	}

	if url, ok := softwareApp["url"].(string); ok {
		parts := strings.Split(strings.TrimSuffix(url, "/"), "/")
		if len(parts) > 0 {
			page.Tool.Slug = parts[len(parts)-1]
		}
	}

	if desc, ok := softwareApp["description"].(string); ok {
		cleanDesc := regexp.MustCompile(`<[^>]+>`).ReplaceAllString(desc, "")
		cleanDesc = strings.ReplaceAll(cleanDesc, "&nbsp;", " ")
		cleanDesc = html.UnescapeString(cleanDesc)
		cleanDesc = strings.TrimSpace(cleanDesc)
		page.Tool.Description = cleanDesc
	}

	if lang, ok := softwareApp["programmingLanguage"].(string); ok {
		page.Tool.Language = lang
	}

	if repo, ok := softwareApp["codeRepository"].(string); ok {
		page.Tool.CodeRepository = repo
	}

	if date, ok := softwareApp["datePublished"].(string); ok {
		page.Tool.DatePublished = date
	}

	titleMatch := titleRegex.FindStringSubmatch(string(data))
	if len(titleMatch) >= 2 {
		titleParts := strings.Split(titleMatch[1], " - ")
		if len(titleParts) > 0 {
			if page.Tool.Name == "" {
				page.Tool.Name = strings.TrimSpace(titleParts[0])
			}
		}
	}

	taglineMatch := taglineRegex.FindStringSubmatch(string(data))
	if len(taglineMatch) >= 2 {
		tagline := strings.TrimSpace(taglineMatch[1])
		tagline = html.UnescapeString(tagline)
		page.Tool.Tagline = tagline
	}

	if installMatch := installRegex.FindStringSubmatch(string(data)); len(installMatch) >= 2 {
		installData := installMatch[1]
		installData = strings.ReplaceAll(installData, "&quot;", "\"")
		installData = strings.ReplaceAll(installData, "&#34;", "\"")
		installData = strings.ReplaceAll(installData, "&amp;", "&")

		var raw map[string]any
		if err := json.Unmarshal([]byte(installData), &raw); err == nil {
			for platform, val := range raw {
				switch v := val.(type) {
				case map[string]any:
					for method, cmd := range v {
						if cmdStr, ok := cmd.(string); ok {
							// Decode HTML entities like &#39; → '
							cmdStr = html.UnescapeString(cmdStr)
							methods := strings.Split(method, "/")
							for _, m := range methods {
								m = strings.TrimSpace(m)
								key := platform + ":" + m
								page.Installations[key] = cmdStr
							}
						}
					}
				case string:
					// Decode HTML entities like &#39; → '
					page.Installations[platform] = html.UnescapeString(v)
				}
			}
		}
	}

	if page.Tool.Slug == "" {
		page.Tool.Slug = strings.ToLower(strings.ReplaceAll(page.Tool.Name, " ", "-"))
	}

	return page, nil
}

// ToInstallInstructions converts the parsed installations to database records.
func (p *DetailPage) ToInstallInstructions() []db.InstallInstruction {
	var insts []db.InstallInstruction
	for platform, command := range p.Installations {
		inst := db.InstallInstruction{
			ID:       uuid.New().String(),
			ToolID:   p.Tool.ID,
			Platform: platform,
			Command:  command,
		}
		insts = append(insts, inst)
	}

	return insts
}

// ToTool returns the parsed tool as a database record.
func (p *DetailPage) ToTool() *db.Tool {
	return &p.Tool
}

// ParseError wraps a parsing error with context.
type ParseError struct {
	Inner error
}

func (e *ParseError) Error() string {
	return e.Inner.Error()
}

func (e *ParseError) Unwrap() error {
	return e.Inner
}

// NewParseError creates a new ParseError with a formatted message.
func NewParseError(format string, args ...any) *ParseError {
	return &ParseError{
		Inner: fmt.Errorf(format, args...),
	}
}
