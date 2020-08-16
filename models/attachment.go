package models

import (
	"archive/zip"
	"bytes"
	"encoding/base64"
	"io"
	"io/ioutil"
	"path/filepath"
	"strings"
)

// Attachment contains the fields and methods for
// an email attachment
type Attachment struct {
	Id             int64  `json:"-"`
	TemplateId     int64  `json:"-"`
	Content        string `json:"content"`
	Type           string `json:"type"`
	Name           string `json:"name"`
	noTemplateVars bool
}

// ApplyTemplate parses different attachment files and applies the supplied phishing template.
func (a *Attachment) ApplyTemplate(ptx PhishingTemplateContext) (io.Reader, error) {

	var processedAttachment string

	decodedAttachment, err := base64.StdEncoding.DecodeString(a.Content) // Decode the attachment
	if err != nil {
		return nil, err
	}

	if a.noTemplateVars == true {
		processedAttachment = string(decodedAttachment)
	} else {

		// Decided to use the file extension rather than the content type, as there seems to be quite
		//  a bit of variability with types. e.g sometimes a Word docx file would have:
		//   "application/vnd.openxmlformats-officedocument.wordprocessingml.document"
		fileExtension := filepath.Ext(a.Name)

		switch fileExtension {

		case ".docx", ".docm", ".pptx", ".xlsx", ".xlsm":
			// Most modern office formats are xml based and can be unarchived.
			// .docm and .xlsm files are comprised of xml, and a binary blob for the macro code

			// Create a new zip reader from the file
			zipReader, err := zip.NewReader(bytes.NewReader(decodedAttachment), int64(len(decodedAttachment)))
			if err != nil {
				return nil, err
			}

			newZipArchive := new(bytes.Buffer)
			zipWriter := zip.NewWriter(newZipArchive) // For writing the new archive

			// i. Read each file from the Word document archive
			// ii. Apply the template to it
			// iii. Add the templated content to a new zip Word archive
			fileContainedTemplatesVars := false
			for _, zipFile := range zipReader.File {
				ff, err := zipFile.Open()
				if err != nil {
					return nil, err
				}
				defer ff.Close()
				contents, err := ioutil.ReadAll(ff)
				if err != nil {
					return nil, err
				}
				subFileExtension := filepath.Ext(zipFile.Name)
				var tFile string
				if subFileExtension == ".xml" || subFileExtension == ".rels" { // Ignore other files, e.g binary ones and images
					// For each file apply the template.
					tFile, err = ExecuteTemplate(string(contents), ptx)
					if err != nil {
						return nil, err
					}
					// Check if the subfile changed. We only need this to be set once to know in the future to check the 'parent' file
					if tFile != string(contents) {
						fileContainedTemplatesVars = true
					}

				} else {
					tFile = string(contents) // Could move this to the declaration of tFile, but might be confusing to read
				}
				// Write new Word archive
				newZipFile, err := zipWriter.Create(zipFile.Name)
				if err != nil {
					zipWriter.Close() // Don't use defer when writing files https://www.joeshaw.org/dont-defer-close-on-writable-files/
					return nil, err
				}
				_, err = newZipFile.Write([]byte(tFile))
				if err != nil {
					zipWriter.Close()
					return nil, err
				}

			}

			// If no files in the archive had template variables, we set the 'parent' file to not be checked in the future
			if fileContainedTemplatesVars == false {
				a.noTemplateVars = true
			}

			zipWriter.Close()
			processedAttachment = newZipArchive.String()

		case ".txt", ".html":
			processedAttachment, err = ExecuteTemplate(string(decodedAttachment), ptx)
			if err != nil {
				return nil, err
			}
		default:
			// We have two options here; either apply template to all files, or none. Probably safer to err on the side of none.
			processedAttachment = string(decodedAttachment) // Option one: Do nothing
			//processedAttachment, err = ExecuteTemplate(string(decodedAttachment), ptx) // Option two: Template all files
		}

		// Check if applying the template altered the file contents. If not, let's not apply the template again to that file.
		// This doesn't work very well with .docx etc files, as the unzipping and rezipping seems to alter them, so those
		// file have their own logic for checking this (above).
		if processedAttachment == string(decodedAttachment) {
			a.noTemplateVars = true
		}

	}

	decoder := strings.NewReader(processedAttachment)
	return decoder, nil

}
