package AzureDevopsClient

import "time"

type IdentifyRef struct {
	Id          string
	DisplayName string
	ProfileUrl  string
	UniqueName  string
	Url         string
	Descriptor  string
}

type AgentPoolQueue struct {
	Id   int64
	Name string
	Pool AgentPool
	Url  string
}

type AgentPool struct {
	Id       int64
	IsHosted bool
	Name     string
}

type Link struct {
	Href string
}

type Links struct {
	Self     Link
	Web      Link
	Source   Link
	Timeline Link
	Badge    Link
}

type Author struct {
	Name  string
	Email string
	Date  time.Time
}
