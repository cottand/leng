package main

import (
	"strings"
	"testing"
)

func TestRecordsGroupByHost(t *testing.T) {
	recordsString := []string{
		"example.com          IN  A       10.10.0.1 ",
		"boo.org              IN  A       10.10.0.2 ",
		"boo.org              IN  A       10.10.0.3 ",
	}
	records := NewCustomDNSRecordsFromText(recordsString)

	if len(records) != 2 {
		t.Fatalf("map should contain 2 hosts, but had %v", records)
	}

	var ex *CustomDNSRecords
	for _, record := range records {
		if strings.Contains(record.name, "example.com") {
			ex = &record
		}
	}
	if ex == nil {
		t.Fatalf("map should contain example.com")
	}
	if rrs := ex.answer; len(rrs) != 1 &&
		strings.Contains(rrs[0].String(), "10.10.0.1") &&
		rrs[0].Header().Name != "example.com" {
		t.Fatalf("should have 1 answer -> 10.10.0.1, but had %v", rrs)
	}

	var boo *CustomDNSRecords
	for _, record := range records {
		if strings.Contains(record.name, "boo.org") {
			boo = &record
		}
	}

	if boo == nil {
		t.Fatalf("map should contain boo.org, but is %v", boo.name)
	}
	if rrs := ex.answer; len(rrs) != 2 &&
		strings.Contains(rrs[0].String(), "10.10.0.2") &&
		strings.Contains(rrs[1].String(), "10.10.0.3") &&
		rrs[0].Header().Name != "boo.org" &&
		rrs[1].Header().Name != "boo.org" {
		t.Fatalf("should have 1 answer -> 10.10.0.1, but had %v", rrs)
	}
}
