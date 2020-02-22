package script

import (
	"fmt"
	"mikrotik-hosts-parser/hostsfile"
	hostsParser "mikrotik-hosts-parser/hostsfile/parser"
	"mikrotik-hosts-parser/settings/serve"
	"net/http"
)

type sourceResponse struct {
	URL      string
	Response *http.Response
	Error    error
}

type hostsRecords struct {
	Records []*hostsfile.Record
}

// RouterOsScriptSourceGenerationHandlerFunc generates RouterOS script source and writes it response.
func RouterOsScriptSourceGenerationHandlerFunc(serveSettings *serve.Settings) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		queryParameters, queryErr := newQueryParametersBagUsingQueryValues(r.URL.Query())

		// Validate query parameters parsing
		if queryErr != nil {
			w.WriteHeader(http.StatusBadRequest)
			_, _ = w.Write([]byte("## Query parameters error: " + queryErr.Error()))

			return
		}

		var (
			processingErrors       = make([]string, 0) // stack with processing errors
			sourcesLen             = len(queryParameters.SourceUrls)
			sourceResponsesChannel = make(chan sourceResponse, sourcesLen) // channel for source responses
		)

		// fetch sources async and write responses into channel
		for _, sourceUrl := range queryParameters.SourceUrls {
			go func(sourceUrl string, maxLength int) {
				fmt.Println(sourceUrl, maxLength)
				// do request
				response, err := fetchSourceContent(sourceUrl, maxLength)
				// send request result into channel
				sourceResponsesChannel <- sourceResponse{URL: sourceUrl, Response: response, Error: err}
			}(sourceUrl, serveSettings.RouterScript.MaxSourceSize)
		}

		var (
			parser       = hostsParser.NewParser()
			hostsRecords = make([]*hostsfile.Record, 0) // hosts records stack
		)

		// read source responses and pass it into hosts file parser
		for i := 0; i < sourcesLen; i++ {
			resp := <-sourceResponsesChannel
			// if response contains error - skip it
			if resp.Error != nil {
				processingErrors = append(processingErrors, resp.URL+": "+resp.Error.Error())
				continue
			}

			// parse response content
			records, parseErr := parser.Parse(resp.Response.Body)
			_ = resp.Response.Body.Close()
			if parseErr != nil {
				processingErrors = append(processingErrors, resp.URL+": "+parseErr.Error())
			}

			// and append results into hosts records stack
			hostsRecords = append(hostsRecords, records...)
		}

		// close responses channels after all
		close(sourceResponsesChannel)

		//var (
		//	mikroticDnsEntries = make()
		//)

		// @todo: convert `[]*hostsfile.Record` into microtic static entries with duplicates removal

		//time.Sleep(time.Second * 3)
		fmt.Println(processingErrors)
		//for _, record := range hostsRecords {
		//	fmt.Println(record)
		//}
	}
}

func (records *hostsRecords) removeDuplicates() {
	//keys := make(map[string]*hostsfile.Record)
	//
	//for _, record := range records.Records {
	//	if _, ok := keys[record.Hosts]; !ok {
	//
	//	}
	//}
}
