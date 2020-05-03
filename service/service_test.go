package service

import (
	"reflect"
	"testing"
)

func TestGetService(t *testing.T) {
	type testSample struct {
		svcName string
		svc     Service
	}

	sampleTable := []testSample{
		{
			svcName: "virustotal",
			svc:     NewVirusTotalService(),
		},
		{
			svcName: "vmray",
			svc:     NewVmrayService(),
		},
		{
			svcName: "metadefender",
			svc:     NewMetaDefenderService(),
		},
		{
			svcName: "somename",
			svc:     nil,
		},
	}

	for _, sample := range sampleTable {
		svc := GetService(sample.svcName)

		gotType := reflect.TypeOf(svc)
		wantType := reflect.TypeOf(sample.svc)
		if gotType != wantType {
			t.Errorf("GetService failed for %s , wanted: %v , got: %v", sample.svcName, wantType, gotType)
		}
	}

}

func TestLocalService(t *testing.T) {
	type testSample struct {
		svcName string
		svc     LocalService
	}

	sampleTable := []testSample{
		{
			svcName: "clamav",
			svc:     NewClamavService(),
		},
		{
			svcName: "somename",
			svc:     nil,
		},
	}

	for _, sample := range sampleTable {
		svc := GetLocalService(sample.svcName)

		gotType := reflect.TypeOf(svc)
		wantType := reflect.TypeOf(sample.svc)
		if gotType != wantType {
			t.Errorf("GetLocalService failed for %s , wanted: %v , got: %v", sample.svcName, wantType, gotType)
		}
	}
}

func TestIsServiceLocal(t *testing.T) {
	type testSample struct {
		svcName string
		isLocal bool
	}

	sampleTable := []testSample{
		{
			svcName: "clamav",
			isLocal: true,
		},
		{
			svcName: "virustotal",
			isLocal: false,
		},
		{
			svcName: "somename",
			isLocal: false,
		},
	}

	for _, sample := range sampleTable {
		got := IsServiceLocal(sample.svcName)
		want := sample.isLocal

		if got != want {
			t.Errorf("IsServiceLocal failed for %s, wanted: %v , got: %v", sample.svcName, want, got)
		}

	}
}
