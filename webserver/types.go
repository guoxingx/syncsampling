package webserver

import (
	"sort"
	"strconv"
)

type Config struct {
	Interval int `json:"interval"`
}

type Action uint8

const (
	_ Action = iota
	ActionStart
	ActionReady
	ActionStop
	ActionPause
	ActionContinue
)

func sortFiles(s []string) []string {
	sf := sortedFiles(s)
	sort.Sort(sf)
	return sf
}

type sortedFiles []string

func (s sortedFiles) Len() int { return len(s) }
func (s sortedFiles) Less(i, j int) bool {
	li := len(s[i])
	lj := len(s[j])

	nsi := s[i][2 : li-4]
	nsj := s[j][2 : lj-4]

	ni, _ := strconv.ParseInt(nsi, 10, 64)
	nj, _ := strconv.ParseInt(nsj, 10, 64)

	return ni < nj
}
func (s sortedFiles) Swap(i, j int) { s[i], s[j] = s[j], s[i] }
