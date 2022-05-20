package main

import (
	"encoding/xml"
	"fmt"
	"io/ioutil"
	"os"
)

const ( // password types
	RAW  string = "raw"
	NTLM string = "ntml"
)

type CredentialDB struct {
	Credentials []Credential `xml:"credentials"`
}

type Credential struct {
	Username     string `xml:"user"`
	Password     string `xml:"password"`
	PasswordType string `xml:"passwordType"`
	Domain       string `xml:"domain"`
}

func (db *CredentialDB) credExists(cred Credential) bool {
	for i := 0; i < len(db.Credentials); i++ {
		if db.Credentials[i] == cred {
			return true
		}
	}
	return false
}

func (db *CredentialDB) writeToDisk() {
	file, err := xml.MarshalIndent(db, "", "  ")

	if err != nil {
		fmt.Println("Failed to marshall credential database!")
		return
	}

	err = ioutil.WriteFile("db/creds.xml", file, 0644)

	if err != nil {
		fmt.Println("Failed to write credentials to disk")
	}
}

func (db *CredentialDB) loadCreds() {
	xmlDbFile, err := os.Open("db/creds.xml")

	if err != nil {
		fmt.Println("Couldn't load db/creds.xml")
		return
	}

	defer xmlDbFile.Close()
	byteValue, _ := ioutil.ReadAll(xmlDbFile)
	xml.Unmarshal(byteValue, db)
	fmt.Println("Loaded harvested credentials from disk")
}

func (db *CredentialDB) addCredential(cred Credential) {
	if !db.credExists(cred) {
		db.Credentials = append(db.Credentials, cred)
		db.writeToDisk()
	}
}

func (db *CredentialDB) getDomainCreds(domain string) []Credential {
	var results []Credential
	for i := 0; i < len(db.Credentials); i++ {
		if db.Credentials[i].Domain == domain {
			results = append(results, db.Credentials[i])
		}
	}
	return results
}

func (db *CredentialDB) getUsernameCreds(username string) []Credential {
	var results []Credential
	for i := 0; i < len(db.Credentials); i++ {
		if db.Credentials[i].Username == username {
			results = append(results, db.Credentials[i])
		}
	}
	return results
}
