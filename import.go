// Copyright 2016 Takbok, Inc. All rights reserved.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
// http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package demo

import (
	"bufio"
	"bytes"
	"encoding/csv"
	"fmt"
	"log"
	"net/http"
	"strings"

	"appengine"

	newappengine "google.golang.org/appengine"
)

func init() {
	http.HandleFunc("/import", handleImport)
	http.HandleFunc("/import/do", handleImportDo)
}

func handleImport(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		http.Redirect(w, r, "/?error=noDirectAccess", http.StatusTemporaryRedirect)
		return
	}

	url, err := getProperDomainNameFromUrl(r.FormValue("url"))
	if err != nil {
		http.Redirect(w, r, "/?error=badUrl", http.StatusTemporaryRedirect)
		return
	}

	if !isUrlOnGoogleApp(w, r, url) {
		http.Redirect(w, r, "/?error=notOnGoogleApps", http.StatusTemporaryRedirect)
		return
	}

	r.ParseMultipartForm(32 << 20)
	inp_file, _, err = r.FormFile("inputfile")
	if err != nil {
		log.Print("\n returning bcoz of error 1")
		log.Print(err)
		return
	}
	defer inp_file.Close()

	x := AppState{url}
	ctx := appengine.NewContext(r)
	config.RedirectURL = fmt.Sprintf(`http://%s/import/do`, r.Host)

	url = config.AuthCodeURL(x.encodeState())
	ctx.Infof("Auth: %v", url)

	http.Redirect(w, r, url, http.StatusTemporaryRedirect)
}

func handleImportDo(w http.ResponseWriter, r *http.Request) {
	y := r.FormValue("state")

	state := new(AppState)
	state.decodeState(y)

	ctx := appengine.NewContext(r)
	newctx := newappengine.NewContext(r)

	tok, err := config.Exchange(newctx, r.FormValue("code"))
	if err != nil {
		ctx.Errorf("exchange error: %v", err)
		return
	}

	client := config.Client(newctx, tok)

	cr := csv.NewReader(bufio.NewReader(inp_file))
	records, err := cr.ReadAll()
	if err != nil {
		log.Print("\n CSV file error")
		ctx.Errorf("%v", err)
		return
	}

	names := records[0]
	datalen := len(records)
	log.Print("\n Loop started")

	for i := 1; i < datalen; i++ {
		rec := records[i]
		buf := new(bytes.Buffer)
		fmt.Fprintf(buf, `<atom:entry xmlns:atom='http://www.w3.org/2005/Atom' xmlns:gd='http://schemas.google.com/g/2005'>
<atom:category scheme='http://schemas.google.com/g/2005#kind' term='http://schemas.google.com/contact/2008#contact' />
<atom:content type='text'>Notes</atom:content>
`)
		var nameBuf, emailBuf, imBuf, orgBuf, extendedBuf string
		//var phoneBuf string
		orgBuf = `<gd:organization rel="http://schemas.google.com/g/2005#work" primary="true">` + "\n"
		numChk, maxChk := 0, 9
		for j, s := range names {
			if s == "E-mail Address" {
				emailBuf += fmt.Sprintf(`<gd:email rel="http://schemas.google.com/g/2005#home" address="%v" primary="true"/>` + "\n", rec[j])
				continue
			}
			if strings.Contains(s, "E-mail") {
				emailBuf += fmt.Sprintf(`<gd:email rel="http://schemas.google.com/g/2005#other" address="%v"/>` + "\n", rec[j])
				continue
			}
			if strings.Contains(s, "IM") {
				imBuf += fmt.Sprintf(`<gd:im address="%v" rel="http://schemas.google.com/g/2005#other"/>` + "\n", rec[j])
				continue
			}
		
			switch(s){
				case "Name" :
				nameBuf += fmt.Sprintf("\n" + `<gd:fullName>%v</gd:fullName>`, rec[j])
				break
				case "GivenName" :
				nameBuf += fmt.Sprintf("\n" + `<gd:givenName>%v</gd:givenName>`, rec[j])
				break
				case "FamilyName" :
				nameBuf += fmt.Sprintf("\n" + `<gd:familyName>%v</gd:familyName>`, rec[j])
				break;
				/*case "Company" :
				orgBuf += fmt.Sprintf(`<gd:orgName>%v</gd:orgName>` + "\n", rec[j])
				break
				case "Job Title" :
				orgBuf += fmt.Sprintf(`<gd:orgTitle>%v</gd:orgTitle>` + "\n", rec[j])
				break
				case "Department" :
				orgBuf += fmt.Sprintf(`<gd:orgDepartment>%v</gd:orgDepartment>` + "\n", rec[j])
				break
				case "Job Description" :
				orgBuf += fmt.Sprintf(`<gd:orgJobDescription>%v</gd:orgJobDescription>` + "\n", rec[j])
				break
				case "Business Fax" :
				phoneBuf += fmt.Sprintf(`<gd:phoneNumber rel="http://schemas.google.com/g/2005#work_fax" label="%v">\n%v\n</gd:phoneNumber>\n`, s, rec[j])
				break;
				case "Business Phone" :
				phoneBuf += fmt.Sprintf(`<gd:phoneNumber rel="http://schemas.google.com/g/2005#work" label="%v">\n%v\n</gd:phoneNumber>\n`, s, rec[j])
				break;
				case "Business Phone 2" :
				phoneBuf += fmt.Sprintf(`<gd:phoneNumber rel="http://schemas.google.com/g/2005#other" label="%v">\n%v\n</gd:phoneNumber>\n`, s, rec[j])
				break;
				case "Home Fax" :
				phoneBuf += fmt.Sprintf(`<gd:phoneNumber rel="http://schemas.google.com/g/2005#home_fax" label="%v">\n%v\n</gd:phoneNumber>\n`, s, rec[j])
				break;
				case "Home Phone" :
				phoneBuf += fmt.Sprintf(`<gd:phoneNumber rel="http://schemas.google.com/g/2005#home" label="%v">\n%v\n</gd:phoneNumber>\n`, s, rec[j])
				break;
				case "Home Phone 2" :
				phoneBuf += fmt.Sprintf(`<gd:phoneNumber rel="http://schemas.google.com/g/2005#other" label="%v">\n%v\n</gd:phoneNumber>\n`, s, rec[j])
				break;
				case "Other Phone" :
				phoneBuf += fmt.Sprintf(`<gd:phoneNumber rel="http://schemas.google.com/g/2005#other" label="%v">'%v\n</gd:phoneNumber>\n`, s, rec[j])
				break;
				case "Mobile Phone" :
				phoneBuf += fmt.Sprintf(`<gd:phoneNumber rel="http://schemas.google.com/g/2005#mobile" label="%v">\n%v\n</gd:phoneNumber>\n`, s, rec[j])
				break;
				case "Pager" :
				phoneBuf += fmt.Sprintf(`<gd:phoneNumber rel="http://schemas.google.com/g/2005#pager" label="%v">\n%v\n</gd:phoneNumber>\n`, s, rec[j])
				break;*/
				default :
				if numChk > maxChk {
					continue
				}
				extendedBuf += fmt.Sprintf(`<gd:extendedProperty name="%v" value="%v" />` + "\n", s, rec[j])
				numChk += 1
				break
			}
		}
		nameBuf = "<gd:name>" + nameBuf + "</gd:name>\n"

		fmt.Fprintf(buf, nameBuf)
		fmt.Fprintf(buf, emailBuf)
		fmt.Fprintf(buf, imBuf)
		orgBuf += "</gd:organization>\n"
		//fmt.Fprintf(buf, orgBuf)
		//fmt.Fprintf(buf, phoneBuf)
		fmt.Fprintf(buf, extendedBuf)

		fmt.Fprintf(buf, `</atom:entry>`)

		res, _ := client.Post(fmt.Sprintf(feedUrl, state.Domain), `application/atom+xml`, strings.NewReader(buf.String()))

		fmt.Fprintf(w, "Result: %v<br/>", res.Status)

	}
}
