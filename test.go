package main

import (
	"flag"
	"log"
)

const (
	Test = "teset"
	Ca   = "ca"
	Baz  = "asd"
	AC   = "ac"
	B    = "asd"
	C    = "asdas"
)

func main() {
	jiraUserFlag := flag.String("user", "", "JIRA user name, eg: devops@manabie.com")
	jiraTokenFlag := flag.String("token", "", "JIRA user token")
	log.Println("jiraUserFlag", jiraUserFlag)
	log.Println("jiraTokenFlag", jiraTokenFlag)

}
