# Troveler

The idea is to have a kind of a local copy of terminaltrove.com. github.com/charmbracelet libraries are always nice. spf13/cobra will also come in useful
The page is very accomodating as it just hands us a JSON if we query it naively, e.g.: 'https://terminaltrove.com/search?q=*&page=1&per_page=100'
We can also filter in out search, e.g.: https://terminaltrove.com/search?q=*&page=1&per_page=8&filter_by=%28operating_systems%3A%3D%5B%60linux%60%5D%29

The result of the first URL can be found in `test.json`
Knowledge gained so far:
  - 100 is the max per page, higher values will still return 100 results
  - .facet_counts[].field_name contains filterable fields (c.f. 2nd URL with the filter_by param):
    - language
    - license
    - tool_of_the_week
    - operating_systems
  - .facet_counts[].counts[n].value contains the valid values for filter field n
  - .found shows the total number of hits leading us to know how many pages there are: math.Ceil(found / 100), currently 807 tools in total meaning 9 pages a 100 records
    - .out_of seems to contain the total number of tools in their db
  - Details can be found under https://terminaltrove.com/$slug/ e.g. https://terminaltrove.com/aichat/
    - There is a <script type="application/ld+json"> section that contains metadata (see inside.html / inside_pretty.html)
    - There's a div with id="install" with a data-install tag that contains ways to install it for different systems/tools (see aichat.html / aichat_pretty.html)

Goal: create a searchable copy by crawling the results and storing that data in a db (sqlite for now but adapter should be exchangable).
I don't know if the date_create(_ts) or datePublished in the <script> changes if the content changes - would be nice if so, then a full update wouldn't require downloading everthing
So: one command for updating everything (crawl all search pages), one command for searching (the offline db, search in name and tagline) - generate a table showing name|tagline|language (make it pretty with charmbracelet)
The install instructions are also rather interesting but that's step 2.
