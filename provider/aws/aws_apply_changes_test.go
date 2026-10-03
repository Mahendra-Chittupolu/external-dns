/*
Copyright 2017 The Kubernetes Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package aws

import (
	"context"
	"testing"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	route53types "github.com/aws/aws-sdk-go-v2/service/route53/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"sigs.k8s.io/external-dns/provider/blueprint"

	"sigs.k8s.io/external-dns/endpoint"
	"sigs.k8s.io/external-dns/plan"
	"sigs.k8s.io/external-dns/provider"
)

func TestAWSApplyChanges(t *testing.T) {
	tests := []struct {
		name       string
		setup      func(p *AWSProvider) context.Context
		listRRSets int
	}{
		{"no cache", func(_ *AWSProvider) context.Context { return t.Context() }, 0},
		{"cached", func(p *AWSProvider) context.Context {
			ctx := t.Context()
			records, err := p.Records(ctx)
			require.NoError(t, err)
			return context.WithValue(ctx, provider.RecordsContextKey, records)
		}, 0},
	}

	for _, tt := range tests {
		provider, _ := newAWSProvider(t, endpoint.NewDomainFilter([]string{"ext-dns-test-2.teapot.zalan.do."}), provider.NewZoneIDFilter([]string{}), provider.NewZoneTypeFilter(""), defaultEvaluateTargetHealth, false, false, []route53types.ResourceRecordSet{
			{
				Name:            aws.String("update-test.zone-1.ext-dns-test-2.teapot.zalan.do."),
				Type:            route53types.RRTypeA,
				TTL:             aws.Int64(defaultTTL),
				ResourceRecords: []route53types.ResourceRecord{{Value: aws.String("8.8.8.8")}},
			},
			{
				Name:            aws.String("update-test-aaaa.zone-1.ext-dns-test-2.teapot.zalan.do."),
				Type:            route53types.RRTypeAaaa,
				TTL:             aws.Int64(defaultTTL),
				ResourceRecords: []route53types.ResourceRecord{{Value: aws.String("2606:4700:4700::1111")}},
			},
			{
				Name:            aws.String("delete-test.zone-1.ext-dns-test-2.teapot.zalan.do."),
				Type:            route53types.RRTypeA,
				TTL:             aws.Int64(defaultTTL),
				ResourceRecords: []route53types.ResourceRecord{{Value: aws.String("8.8.8.8")}},
			},
			{
				Name:            aws.String("delete-test-aaaa.zone-1.ext-dns-test-2.teapot.zalan.do."),
				Type:            route53types.RRTypeAaaa,
				TTL:             aws.Int64(defaultTTL),
				ResourceRecords: []route53types.ResourceRecord{{Value: aws.String("2606:4700:4700::1111")}},
			},
			{
				Name:            aws.String("update-test.zone-2.ext-dns-test-2.teapot.zalan.do."),
				Type:            route53types.RRTypeA,
				TTL:             aws.Int64(defaultTTL),
				ResourceRecords: []route53types.ResourceRecord{{Value: aws.String("8.8.4.4")}},
			},
			{
				Name:            aws.String("update-test-aaaa.zone-2.ext-dns-test-2.teapot.zalan.do."),
				Type:            route53types.RRTypeAaaa,
				TTL:             aws.Int64(defaultTTL),
				ResourceRecords: []route53types.ResourceRecord{{Value: aws.String("2606:4700:4700::1001")}},
			},
			{
				Name:            aws.String("delete-test.zone-2.ext-dns-test-2.teapot.zalan.do."),
				Type:            route53types.RRTypeA,
				TTL:             aws.Int64(defaultTTL),
				ResourceRecords: []route53types.ResourceRecord{{Value: aws.String("8.8.4.4")}},
			},
			{
				Name:            aws.String("delete-test-aaaa.zone-2.ext-dns-test-2.teapot.zalan.do."),
				Type:            route53types.RRTypeAaaa,
				TTL:             aws.Int64(defaultTTL),
				ResourceRecords: []route53types.ResourceRecord{{Value: aws.String("2606:4700:4700::1001")}},
			},
			{
				Name:            aws.String("update-test-a-to-cname.zone-1.ext-dns-test-2.teapot.zalan.do."),
				Type:            route53types.RRTypeA,
				TTL:             aws.Int64(defaultTTL),
				ResourceRecords: []route53types.ResourceRecord{{Value: aws.String("1.1.1.1")}},
			},
			{
				Name: aws.String("update-test-alias-to-cname.zone-1.ext-dns-test-2.teapot.zalan.do."),
				Type: route53types.RRTypeA,
				AliasTarget: &route53types.AliasTarget{
					DNSName:              aws.String("foo.eu-central-1.elb.amazonaws.com."),
					EvaluateTargetHealth: true,
					HostedZoneId:         aws.String("Z215JYRZR1TBD5"),
				},
			},
			{
				Name: aws.String("update-test-alias-to-cname.zone-1.ext-dns-test-2.teapot.zalan.do."),
				Type: route53types.RRTypeAaaa,
				AliasTarget: &route53types.AliasTarget{
					DNSName:              aws.String("foo.eu-central-1.elb.amazonaws.com."),
					EvaluateTargetHealth: true,
					HostedZoneId:         aws.String("Z215JYRZR1TBD5"),
				},
			},
			{
				Name:            aws.String("update-test-cname.zone-1.ext-dns-test-2.teapot.zalan.do."),
				Type:            route53types.RRTypeCname,
				TTL:             aws.Int64(defaultTTL),
				ResourceRecords: []route53types.ResourceRecord{{Value: aws.String("bar.elb.amazonaws.com")}},
			},
			{
				Name: aws.String("update-test-cname-alias.zone-1.ext-dns-test-2.teapot.zalan.do."),
				Type: route53types.RRTypeA,
				AliasTarget: &route53types.AliasTarget{
					DNSName:              aws.String("bar.elb.amazonaws.com."),
					EvaluateTargetHealth: true,
					HostedZoneId:         aws.String("Z215JYRZR1TBD5"),
				},
			},
			{
				Name: aws.String("update-test-cname-alias.zone-1.ext-dns-test-2.teapot.zalan.do."),
				Type: route53types.RRTypeAaaa,
				AliasTarget: &route53types.AliasTarget{
					DNSName:              aws.String("bar.elb.amazonaws.com."),
					EvaluateTargetHealth: true,
					HostedZoneId:         aws.String("Z215JYRZR1TBD5"),
				},
			},
			{
				Name:            aws.String("delete-test-cname.zone-1.ext-dns-test-2.teapot.zalan.do."),
				Type:            route53types.RRTypeCname,
				TTL:             aws.Int64(defaultTTL),
				ResourceRecords: []route53types.ResourceRecord{{Value: aws.String("qux.elb.amazonaws.com")}},
			},
			{
				Name: aws.String("delete-test-cname-alias.zone-1.ext-dns-test-2.teapot.zalan.do."),
				Type: route53types.RRTypeA,
				AliasTarget: &route53types.AliasTarget{
					DNSName:              aws.String("qux.eu-central-1.elb.amazonaws.com."),
					EvaluateTargetHealth: true,
					HostedZoneId:         aws.String("Z215JYRZR1TBD5"),
				},
			},
			{
				Name: aws.String("delete-test-cname-alias.zone-1.ext-dns-test-2.teapot.zalan.do."),
				Type: route53types.RRTypeAaaa,
				AliasTarget: &route53types.AliasTarget{
					DNSName:              aws.String("qux.eu-central-1.elb.amazonaws.com."),
					EvaluateTargetHealth: true,
					HostedZoneId:         aws.String("Z215JYRZR1TBD5"),
				},
			},
			{
				Name:            aws.String("update-test-multiple.zone-2.ext-dns-test-2.teapot.zalan.do."),
				Type:            route53types.RRTypeA,
				TTL:             aws.Int64(defaultTTL),
				ResourceRecords: []route53types.ResourceRecord{{Value: aws.String("8.8.8.8")}, {Value: aws.String("8.8.4.4")}},
			},
			{
				Name:            aws.String("delete-test-multiple.zone-2.ext-dns-test-2.teapot.zalan.do."),
				Type:            route53types.RRTypeA,
				TTL:             aws.Int64(defaultTTL),
				ResourceRecords: []route53types.ResourceRecord{{Value: aws.String("1.2.3.4")}, {Value: aws.String("4.3.2.1")}},
			},
			{
				Name:            aws.String("delete-test-multiple-aaaa.zone-2.ext-dns-test-2.teapot.zalan.do."),
				Type:            route53types.RRTypeAaaa,
				TTL:             aws.Int64(defaultTTL),
				ResourceRecords: []route53types.ResourceRecord{{Value: aws.String("2606:4700:4700::1111")}, {Value: aws.String("2606:4700:4700::1001")}},
			},
			{
				Name:            aws.String("delete-test-geoproximity.zone-2.ext-dns-test-2.teapot.zalan.do."),
				Type:            route53types.RRTypeA,
				TTL:             aws.Int64(defaultTTL),
				ResourceRecords: []route53types.ResourceRecord{{Value: aws.String("1.2.3.4")}},
				SetIdentifier:   aws.String("geoproximity-delete"),
				GeoProximityLocation: &route53types.GeoProximityLocation{
					AWSRegion: aws.String("us-west-2"),
					Bias:      aws.Int32(10),
				},
			},
			{
				Name:            aws.String("update-test-geoproximity.zone-1.ext-dns-test-2.teapot.zalan.do."),
				Type:            route53types.RRTypeA,
				TTL:             aws.Int64(defaultTTL),
				ResourceRecords: []route53types.ResourceRecord{{Value: aws.String("1.2.3.4")}},
				SetIdentifier:   aws.String("geoproximity-update"),
				GeoProximityLocation: &route53types.GeoProximityLocation{
					LocalZoneGroup: aws.String("usw2-lax1-az2"),
				},
			},
			{
				Name:            aws.String("weighted-to-simple.zone-1.ext-dns-test-2.teapot.zalan.do."),
				Type:            route53types.RRTypeA,
				TTL:             aws.Int64(defaultTTL),
				ResourceRecords: []route53types.ResourceRecord{{Value: aws.String("1.2.3.4")}},
				SetIdentifier:   aws.String("weighted-to-simple"),
				Weight:          aws.Int64(10),
			},
			{
				Name:            aws.String("simple-to-weighted.zone-1.ext-dns-test-2.teapot.zalan.do."),
				Type:            route53types.RRTypeA,
				TTL:             aws.Int64(defaultTTL),
				ResourceRecords: []route53types.ResourceRecord{{Value: aws.String("1.2.3.4")}},
			},
			{
				Name:            aws.String("policy-change.zone-1.ext-dns-test-2.teapot.zalan.do."),
				Type:            route53types.RRTypeA,
				TTL:             aws.Int64(defaultTTL),
				ResourceRecords: []route53types.ResourceRecord{{Value: aws.String("1.2.3.4")}},
				SetIdentifier:   aws.String("policy-change"),
				Weight:          aws.Int64(10),
			},
			{
				Name:            aws.String("set-identifier-change.zone-1.ext-dns-test-2.teapot.zalan.do."),
				Type:            route53types.RRTypeA,
				TTL:             aws.Int64(defaultTTL),
				ResourceRecords: []route53types.ResourceRecord{{Value: aws.String("1.2.3.4")}},
				SetIdentifier:   aws.String("before"),
				Weight:          aws.Int64(10),
			},
			{
				Name:            aws.String("set-identifier-no-change.zone-1.ext-dns-test-2.teapot.zalan.do."),
				Type:            route53types.RRTypeA,
				TTL:             aws.Int64(defaultTTL),
				ResourceRecords: []route53types.ResourceRecord{{Value: aws.String("1.2.3.4")}},
				SetIdentifier:   aws.String("no-change"),
				Weight:          aws.Int64(10),
			},
			{
				Name:            aws.String("update-test-mx.zone-2.ext-dns-test-2.teapot.zalan.do."),
				Type:            route53types.RRTypeMx,
				TTL:             aws.Int64(defaultTTL),
				ResourceRecords: []route53types.ResourceRecord{{Value: aws.String("10 mailhost2.bar.elb.amazonaws.com")}},
			},
			{
				Name:            aws.String("delete-test-mx.zone-2.ext-dns-test-2.teapot.zalan.do."),
				Type:            route53types.RRTypeMx,
				TTL:             aws.Int64(defaultTTL),
				ResourceRecords: []route53types.ResourceRecord{{Value: aws.String("30 mailhost1.foo.elb.amazonaws.com")}},
			},
			{
				Name:            aws.String(specialCharactersEscape("escape-%!s(<nil>)-codes.zone-2.ext-dns-test-2.teapot.zalan.do.")),
				Type:            route53types.RRTypeA,
				TTL:             aws.Int64(defaultTTL),
				ResourceRecords: []route53types.ResourceRecord{{Value: aws.String("1.2.3.4")}},
				SetIdentifier:   aws.String("no-change"),
				Weight:          aws.Int64(10),
			},
			{
				Name: aws.String("delete-test-naptr.zone-2.ext-dns-test-2.teapot.zalan.do."),
				Type: route53types.RRTypeNaptr,
				TTL:  aws.Int64(defaultTTL),
				ResourceRecords: []route53types.ResourceRecord{
					{Value: aws.String(`10 "U" "SIP+DTU" "" _sip._udp.sip1.example.com.`)},
				},
			},
			{
				Name:            aws.String("update-test-naptr.zone-2.ext-dns-test-2.teapot.zalan.do."),
				Type:            route53types.RRTypeNaptr,
				TTL:             aws.Int64(defaultTTL),
				ResourceRecords: []route53types.ResourceRecord{{Value: aws.String(`10 "U" "SIP+DTU" "" _sip._udp.sip1.example.com.`)}},
			},
		})

		createRecords := []*endpoint.Endpoint{
			endpoint.NewEndpoint("create-test.zone-1.ext-dns-test-2.teapot.zalan.do", endpoint.RecordTypeA, "8.8.8.8"),
			endpoint.NewEndpoint("create-test.zone-2.ext-dns-test-2.teapot.zalan.do", endpoint.RecordTypeA, "8.8.4.4"),
			endpoint.NewEndpoint("create-test-aaaa.zone-1.ext-dns-test-2.teapot.zalan.do", endpoint.RecordTypeAAAA, "2606:4700:4700::1111"),
			endpoint.NewEndpoint("create-test-aaaa.zone-2.ext-dns-test-2.teapot.zalan.do", endpoint.RecordTypeAAAA, "2606:4700:4700::1001"),
			endpoint.NewEndpoint("create-test-cname.zone-1.ext-dns-test-2.teapot.zalan.do", endpoint.RecordTypeCNAME, "foo.elb.amazonaws.com"),
			endpoint.NewEndpoint("create-test-cname-alias.zone-1.ext-dns-test-2.teapot.zalan.do", endpoint.RecordTypeCNAME, "foo.elb.amazonaws.com"),
			endpoint.NewEndpoint("create-test-multiple.zone-2.ext-dns-test-2.teapot.zalan.do", endpoint.RecordTypeA, "8.8.8.8", "8.8.4.4"),
			endpoint.NewEndpoint("create-test-multiple-aaaa.zone-2.ext-dns-test-2.teapot.zalan.do", endpoint.RecordTypeAAAA, "2606:4700:4700::1111", "2606:4700:4700::1001"),
			endpoint.NewEndpoint("create-test-mx.zone-1.ext-dns-test-2.teapot.zalan.do", endpoint.RecordTypeMX, "10 mailhost1.foo.elb.amazonaws.com"),
			endpoint.NewEndpoint("create-test-naptr.zone-1.ext-dns-test-2.teapot.zalan.do", endpoint.RecordTypeNAPTR, `10 "U" "SIP+DTU" "" _sip._udp.sip1.example.com.`),
			endpoint.NewEndpoint("create-test-geoproximity-region.zone-1.ext-dns-test-2.teapot.zalan.do", endpoint.RecordTypeA, "8.8.8.8").
				WithSetIdentifier("geoproximity-region").
				WithProviderSpecific(providerSpecificGeoProximityLocationAWSRegion, "us-west-2").
				WithProviderSpecific(providerSpecificGeoProximityLocationBias, "10"),
			endpoint.NewEndpoint("create-test-geoproximity-coordinates.zone-1.ext-dns-test-2.teapot.zalan.do", endpoint.RecordTypeA, "8.8.8.8").
				WithSetIdentifier("geoproximity-coordinates").
				WithProviderSpecific(providerSpecificGeoProximityLocationCoordinates, "60,60"),
		}

		currentRecords := []*endpoint.Endpoint{
			endpoint.NewEndpoint("update-test.zone-1.ext-dns-test-2.teapot.zalan.do", endpoint.RecordTypeA, "8.8.8.8"),
			endpoint.NewEndpoint("update-test.zone-2.ext-dns-test-2.teapot.zalan.do", endpoint.RecordTypeA, "8.8.4.4"),
			endpoint.NewEndpoint("update-test-aaaa.zone-1.ext-dns-test-2.teapot.zalan.do", endpoint.RecordTypeAAAA, "2606:4700:4700::1111"),
			endpoint.NewEndpoint("update-test-aaaa.zone-2.ext-dns-test-2.teapot.zalan.do", endpoint.RecordTypeAAAA, "2606:4700:4700::1001"),
			endpoint.NewEndpoint("update-test-a-to-cname.zone-1.ext-dns-test-2.teapot.zalan.do", endpoint.RecordTypeA, "1.1.1.1"),
			endpoint.NewEndpoint("update-test-alias-to-cname.zone-1.ext-dns-test-2.teapot.zalan.do", endpoint.RecordTypeA, "foo.eu-central-1.elb.amazonaws.com").WithAliasProperty(endpoint.AliasTrue),
			endpoint.NewEndpoint("update-test-alias-to-cname.zone-1.ext-dns-test-2.teapot.zalan.do", endpoint.RecordTypeAAAA, "foo.eu-central-1.elb.amazonaws.com").WithAliasProperty(endpoint.AliasTrue),
			endpoint.NewEndpoint("update-test-cname.zone-1.ext-dns-test-2.teapot.zalan.do", endpoint.RecordTypeCNAME, "bar.elb.amazonaws.com"),
			endpoint.NewEndpoint("update-test-cname-alias.zone-1.ext-dns-test-2.teapot.zalan.do", endpoint.RecordTypeA, "bar.elb.amazonaws.com").WithAliasProperty(endpoint.AliasTrue),
			endpoint.NewEndpoint("update-test-cname-alias.zone-1.ext-dns-test-2.teapot.zalan.do", endpoint.RecordTypeAAAA, "bar.elb.amazonaws.com").WithAliasProperty(endpoint.AliasTrue),
			endpoint.NewEndpoint("update-test-multiple.zone-2.ext-dns-test-2.teapot.zalan.do", endpoint.RecordTypeA, "8.8.8.8", "8.8.4.4"),
			endpoint.NewEndpoint("update-test-multiple-aaaa.zone-2.ext-dns-test-2.teapot.zalan.do", endpoint.RecordTypeAAAA, "2606:4700:4700::1111", "2606:4700:4700::1001"),
			endpoint.NewEndpoint("update-test-geoproximity.zone-1.ext-dns-test-2.teapot.zalan.do", endpoint.RecordTypeA, "1.2.3.4").
				WithSetIdentifier("geoproximity-update").
				WithProviderSpecific(providerSpecificGeoProximityLocationLocalZoneGroup, "usw2-lax1-az2"),
			endpoint.NewEndpoint("weighted-to-simple.zone-1.ext-dns-test-2.teapot.zalan.do", endpoint.RecordTypeA, "1.2.3.4").WithSetIdentifier("weighted-to-simple").WithProviderSpecific(providerSpecificWeight, "10"),
			endpoint.NewEndpoint("simple-to-weighted.zone-1.ext-dns-test-2.teapot.zalan.do", endpoint.RecordTypeA, "1.2.3.4"),
			endpoint.NewEndpoint("policy-change.zone-1.ext-dns-test-2.teapot.zalan.do", endpoint.RecordTypeA, "1.2.3.4").WithSetIdentifier("policy-change").WithProviderSpecific(providerSpecificWeight, "10"),
			endpoint.NewEndpoint("set-identifier-change.zone-1.ext-dns-test-2.teapot.zalan.do", endpoint.RecordTypeA, "1.2.3.4").WithSetIdentifier("before").WithProviderSpecific(providerSpecificWeight, "10"),
			endpoint.NewEndpoint("set-identifier-no-change.zone-1.ext-dns-test-2.teapot.zalan.do", endpoint.RecordTypeA, "1.2.3.4").WithSetIdentifier("no-change").WithProviderSpecific(providerSpecificWeight, "10"),
			endpoint.NewEndpoint("update-test-mx.zone-2.ext-dns-test-2.teapot.zalan.do", endpoint.RecordTypeMX, "10 mailhost2.bar.elb.amazonaws.com"),
			endpoint.NewEndpoint("update-test-naptr.zone-2.ext-dns-test-2.teapot.zalan.do", endpoint.RecordTypeNAPTR, `10 "U" "SIP+DTU" "" _sip._udp.sip1.example.com.`),
			endpoint.NewEndpoint("escape-%!s(<nil>)-codes.zone-1.ext-dns-test-2.teapot.zalan.do", endpoint.RecordTypeA, "1.2.3.4").WithSetIdentifier("policy-change").WithSetIdentifier("no-change").WithProviderSpecific(providerSpecificWeight, "10"),
		}
		updatedRecords := []*endpoint.Endpoint{
			endpoint.NewEndpoint("update-test.zone-1.ext-dns-test-2.teapot.zalan.do", endpoint.RecordTypeA, "1.2.3.4"),
			endpoint.NewEndpoint("update-test.zone-2.ext-dns-test-2.teapot.zalan.do", endpoint.RecordTypeA, "4.3.2.1"),
			endpoint.NewEndpoint("update-test-aaaa.zone-1.ext-dns-test-2.teapot.zalan.do", endpoint.RecordTypeAAAA, "2606:4700:4700::1001"),
			endpoint.NewEndpoint("update-test-aaaa.zone-2.ext-dns-test-2.teapot.zalan.do", endpoint.RecordTypeAAAA, "2606:4700:4700::1111"),
			endpoint.NewEndpoint("update-test-a-to-cname.zone-1.ext-dns-test-2.teapot.zalan.do", endpoint.RecordTypeA, "foo.elb.amazonaws.com").WithAliasProperty(endpoint.AliasTrue),
			endpoint.NewEndpoint("update-test-a-to-cname.zone-1.ext-dns-test-2.teapot.zalan.do", endpoint.RecordTypeAAAA, "foo.elb.amazonaws.com").WithAliasProperty(endpoint.AliasTrue),
			endpoint.NewEndpoint("update-test-alias-to-cname.zone-1.ext-dns-test-2.teapot.zalan.do", endpoint.RecordTypeCNAME, "my-internal-host.example.com"),
			endpoint.NewEndpoint("update-test-cname.zone-1.ext-dns-test-2.teapot.zalan.do", endpoint.RecordTypeCNAME, "baz.elb.amazonaws.com"),
			endpoint.NewEndpoint("update-test-cname-alias.zone-1.ext-dns-test-2.teapot.zalan.do", endpoint.RecordTypeA, "baz.elb.amazonaws.com").WithAliasProperty(endpoint.AliasTrue),
			endpoint.NewEndpoint("update-test-cname-alias.zone-1.ext-dns-test-2.teapot.zalan.do", endpoint.RecordTypeAAAA, "baz.elb.amazonaws.com").WithAliasProperty(endpoint.AliasTrue),
			endpoint.NewEndpoint("update-test-multiple.zone-2.ext-dns-test-2.teapot.zalan.do", endpoint.RecordTypeA, "1.2.3.4", "4.3.2.1"),
			endpoint.NewEndpoint("update-test-multiple-aaaa.zone-2.ext-dns-test-2.teapot.zalan.do", endpoint.RecordTypeAAAA, "2606:4700:4700::1001", "2606:4700:4700::1111"),
			endpoint.NewEndpoint("update-test-geoproximity.zone-1.ext-dns-test-2.teapot.zalan.do", endpoint.RecordTypeA, "1.2.3.4").
				WithSetIdentifier("geoproximity-update").
				WithProviderSpecific(providerSpecificGeoProximityLocationLocalZoneGroup, "usw2-phx2-az1"),
			endpoint.NewEndpoint("weighted-to-simple.zone-1.ext-dns-test-2.teapot.zalan.do", endpoint.RecordTypeA, "1.2.3.4"),
			endpoint.NewEndpoint("simple-to-weighted.zone-1.ext-dns-test-2.teapot.zalan.do", endpoint.RecordTypeA, "1.2.3.4").WithSetIdentifier("simple-to-weighted").WithProviderSpecific(providerSpecificWeight, "10"),
			endpoint.NewEndpoint("policy-change.zone-1.ext-dns-test-2.teapot.zalan.do", endpoint.RecordTypeA, "1.2.3.4").WithSetIdentifier("policy-change").WithProviderSpecific(providerSpecificRegion, "us-east-1"),
			endpoint.NewEndpoint("set-identifier-change.zone-1.ext-dns-test-2.teapot.zalan.do", endpoint.RecordTypeA, "1.2.3.4").WithSetIdentifier("after").WithProviderSpecific(providerSpecificWeight, "10"),
			endpoint.NewEndpoint("set-identifier-no-change.zone-1.ext-dns-test-2.teapot.zalan.do", endpoint.RecordTypeA, "1.2.3.4").WithSetIdentifier("no-change").WithProviderSpecific(providerSpecificWeight, "20"),
			endpoint.NewEndpoint("update-test-mx.zone-2.ext-dns-test-2.teapot.zalan.do", endpoint.RecordTypeMX, "20 mailhost3.foo.elb.amazonaws.com"),
			endpoint.NewEndpoint("update-test-naptr.zone-2.ext-dns-test-2.teapot.zalan.do", endpoint.RecordTypeNAPTR, `20 "U" "SIP+DTU" "" _sip._udp.sip2.example.com.`),
		}

		deleteRecords := []*endpoint.Endpoint{
			endpoint.NewEndpoint("delete-test.zone-1.ext-dns-test-2.teapot.zalan.do", endpoint.RecordTypeA, "8.8.8.8"),
			endpoint.NewEndpoint("delete-test.zone-2.ext-dns-test-2.teapot.zalan.do", endpoint.RecordTypeA, "8.8.4.4"),
			endpoint.NewEndpoint("delete-test-aaaa.zone-1.ext-dns-test-2.teapot.zalan.do", endpoint.RecordTypeAAAA, "2606:4700:4700::1111"),
			endpoint.NewEndpoint("delete-test-aaaa.zone-2.ext-dns-test-2.teapot.zalan.do", endpoint.RecordTypeAAAA, "2606:4700:4700::1001"),
			endpoint.NewEndpoint("delete-test-cname.zone-1.ext-dns-test-2.teapot.zalan.do", endpoint.RecordTypeCNAME, "qux.elb.amazonaws.com"),
			endpoint.NewEndpoint("delete-test-cname-alias.zone-1.ext-dns-test-2.teapot.zalan.do", endpoint.RecordTypeA, "qux.eu-central-1.elb.amazonaws.com").WithAliasProperty(endpoint.AliasTrue),
			endpoint.NewEndpoint("delete-test-cname-alias.zone-1.ext-dns-test-2.teapot.zalan.do", endpoint.RecordTypeAAAA, "qux.eu-central-1.elb.amazonaws.com").WithAliasProperty(endpoint.AliasTrue),
			endpoint.NewEndpoint("delete-test-multiple.zone-2.ext-dns-test-2.teapot.zalan.do", endpoint.RecordTypeA, "1.2.3.4", "4.3.2.1"),
			endpoint.NewEndpoint("delete-test-multiple-aaaa.zone-2.ext-dns-test-2.teapot.zalan.do", endpoint.RecordTypeAAAA, "2606:4700:4700::1111", "2606:4700:4700::1001"),
			endpoint.NewEndpoint("delete-test-geoproximity.zone-2.ext-dns-test-2.teapot.zalan.do", endpoint.RecordTypeA, "1.2.3.4").WithSetIdentifier("geoproximity-delete").WithProviderSpecific(providerSpecificGeoProximityLocationAWSRegion, "us-west-2").WithProviderSpecific(providerSpecificGeoProximityLocationBias, "10"),
			endpoint.NewEndpoint("delete-test-mx.zone-2.ext-dns-test-2.teapot.zalan.do", endpoint.RecordTypeMX, "30 mailhost1.foo.elb.amazonaws.com"),
			endpoint.NewEndpoint("delete-test-naptr.zone-2.ext-dns-test-2.teapot.zalan.do", endpoint.RecordTypeNAPTR, `10 "U" "SIP+DTU" "" _sip._udp.sip1.example.com.`),
		}

		changes := &plan.Changes{
			Create:    createRecords,
			UpdateNew: updatedRecords,
			UpdateOld: currentRecords,
			Delete:    deleteRecords,
		}

		ctx := tt.setup(provider)

		provider.zonesCache = blueprint.NewZoneCache[map[string]*profiledZone](0 * time.Minute)
		counter := NewRoute53APICounter(provider.clients[defaultAWSProfile])
		provider.clients[defaultAWSProfile] = counter
		require.NoError(t, provider.ApplyChanges(ctx, changes))

		assert.Equal(t, 1, counter.calls["ListHostedZonesPages"], tt.name)
		assert.Equal(t, tt.listRRSets, counter.calls["ListResourceRecordSetsPages"], tt.name)

		validateRecords(t, listAWSRecords(t, provider.clients[defaultAWSProfile], "/hostedzone/zone-1.ext-dns-test-2.teapot.zalan.do."), []route53types.ResourceRecordSet{
			{
				Name:            aws.String("create-test.zone-1.ext-dns-test-2.teapot.zalan.do."),
				Type:            route53types.RRTypeA,
				TTL:             aws.Int64(defaultTTL),
				ResourceRecords: []route53types.ResourceRecord{{Value: aws.String("8.8.8.8")}},
			},
			{
				Name:            aws.String("create-test-aaaa.zone-1.ext-dns-test-2.teapot.zalan.do."),
				Type:            route53types.RRTypeAaaa,
				TTL:             aws.Int64(defaultTTL),
				ResourceRecords: []route53types.ResourceRecord{{Value: aws.String("2606:4700:4700::1111")}},
			},
			{
				Name:            aws.String("update-test.zone-1.ext-dns-test-2.teapot.zalan.do."),
				Type:            route53types.RRTypeA,
				TTL:             aws.Int64(defaultTTL),
				ResourceRecords: []route53types.ResourceRecord{{Value: aws.String("1.2.3.4")}},
			},
			{
				Name:            aws.String("update-test-aaaa.zone-1.ext-dns-test-2.teapot.zalan.do."),
				Type:            route53types.RRTypeAaaa,
				TTL:             aws.Int64(defaultTTL),
				ResourceRecords: []route53types.ResourceRecord{{Value: aws.String("2606:4700:4700::1001")}},
			},
			{
				Name: aws.String("update-test-a-to-cname.zone-1.ext-dns-test-2.teapot.zalan.do."),
				Type: route53types.RRTypeA,
				AliasTarget: &route53types.AliasTarget{
					DNSName:              aws.String("foo.elb.amazonaws.com."),
					EvaluateTargetHealth: true,
					HostedZoneId:         aws.String("zone-1.ext-dns-test-2.teapot.zalan.do."),
				},
			},
			{
				Name: aws.String("update-test-a-to-cname.zone-1.ext-dns-test-2.teapot.zalan.do."),
				Type: route53types.RRTypeAaaa,
				AliasTarget: &route53types.AliasTarget{
					DNSName:              aws.String("foo.elb.amazonaws.com."),
					EvaluateTargetHealth: true,
					HostedZoneId:         aws.String("zone-1.ext-dns-test-2.teapot.zalan.do."),
				},
			},
			{
				Name:            aws.String("update-test-alias-to-cname.zone-1.ext-dns-test-2.teapot.zalan.do."),
				Type:            route53types.RRTypeCname,
				TTL:             aws.Int64(defaultTTL),
				ResourceRecords: []route53types.ResourceRecord{{Value: aws.String("my-internal-host.example.com")}},
			},
			{
				Name:            aws.String("create-test-cname.zone-1.ext-dns-test-2.teapot.zalan.do."),
				Type:            route53types.RRTypeCname,
				TTL:             aws.Int64(defaultTTL),
				ResourceRecords: []route53types.ResourceRecord{{Value: aws.String("foo.elb.amazonaws.com")}},
			},
			{
				Name:            aws.String("update-test-cname.zone-1.ext-dns-test-2.teapot.zalan.do."),
				Type:            route53types.RRTypeCname,
				TTL:             aws.Int64(defaultTTL),
				ResourceRecords: []route53types.ResourceRecord{{Value: aws.String("baz.elb.amazonaws.com")}},
			},
			{
				Name:            aws.String("create-test-cname-alias.zone-1.ext-dns-test-2.teapot.zalan.do."),
				Type:            route53types.RRTypeCname,
				TTL:             aws.Int64(defaultTTL),
				ResourceRecords: []route53types.ResourceRecord{{Value: aws.String("foo.elb.amazonaws.com")}},
			},
			{
				Name: aws.String("update-test-cname-alias.zone-1.ext-dns-test-2.teapot.zalan.do."),
				Type: route53types.RRTypeA,
				AliasTarget: &route53types.AliasTarget{
					DNSName:              aws.String("baz.elb.amazonaws.com."),
					EvaluateTargetHealth: true,
					HostedZoneId:         aws.String("zone-1.ext-dns-test-2.teapot.zalan.do."),
				},
			},
			{
				Name: aws.String("update-test-cname-alias.zone-1.ext-dns-test-2.teapot.zalan.do."),
				Type: route53types.RRTypeAaaa,
				AliasTarget: &route53types.AliasTarget{
					DNSName:              aws.String("baz.elb.amazonaws.com."),
					EvaluateTargetHealth: true,
					HostedZoneId:         aws.String("zone-1.ext-dns-test-2.teapot.zalan.do."),
				},
			},
			{
				Name:            aws.String("weighted-to-simple.zone-1.ext-dns-test-2.teapot.zalan.do."),
				Type:            route53types.RRTypeA,
				TTL:             aws.Int64(defaultTTL),
				ResourceRecords: []route53types.ResourceRecord{{Value: aws.String("1.2.3.4")}},
			},
			{
				Name:            aws.String("simple-to-weighted.zone-1.ext-dns-test-2.teapot.zalan.do."),
				Type:            route53types.RRTypeA,
				TTL:             aws.Int64(defaultTTL),
				ResourceRecords: []route53types.ResourceRecord{{Value: aws.String("1.2.3.4")}},
				SetIdentifier:   aws.String("simple-to-weighted"),
				Weight:          aws.Int64(10),
			},
			{
				Name:            aws.String("policy-change.zone-1.ext-dns-test-2.teapot.zalan.do."),
				Type:            route53types.RRTypeA,
				TTL:             aws.Int64(defaultTTL),
				ResourceRecords: []route53types.ResourceRecord{{Value: aws.String("1.2.3.4")}},
				SetIdentifier:   aws.String("policy-change"),
				Region:          route53types.ResourceRecordSetRegionUsEast1,
			},
			{
				Name:            aws.String("set-identifier-change.zone-1.ext-dns-test-2.teapot.zalan.do."),
				Type:            route53types.RRTypeA,
				TTL:             aws.Int64(defaultTTL),
				ResourceRecords: []route53types.ResourceRecord{{Value: aws.String("1.2.3.4")}},
				SetIdentifier:   aws.String("after"),
				Weight:          aws.Int64(10),
			},
			{
				Name:            aws.String("set-identifier-no-change.zone-1.ext-dns-test-2.teapot.zalan.do."),
				Type:            route53types.RRTypeA,
				TTL:             aws.Int64(defaultTTL),
				ResourceRecords: []route53types.ResourceRecord{{Value: aws.String("1.2.3.4")}},
				SetIdentifier:   aws.String("no-change"),
				Weight:          aws.Int64(20),
			},
			{
				Name:            aws.String("create-test-mx.zone-1.ext-dns-test-2.teapot.zalan.do."),
				Type:            route53types.RRTypeMx,
				TTL:             aws.Int64(defaultTTL),
				ResourceRecords: []route53types.ResourceRecord{{Value: aws.String("10 mailhost1.foo.elb.amazonaws.com")}},
			},
			{
				Name:            aws.String("create-test-naptr.zone-1.ext-dns-test-2.teapot.zalan.do."),
				Type:            route53types.RRTypeNaptr,
				TTL:             aws.Int64(defaultTTL),
				ResourceRecords: []route53types.ResourceRecord{{Value: aws.String(`10 "U" "SIP+DTU" "" _sip._udp.sip1.example.com.`)}},
			},
			{
				Name:            aws.String("create-test-geoproximity-region.zone-1.ext-dns-test-2.teapot.zalan.do."),
				Type:            route53types.RRTypeA,
				TTL:             aws.Int64(defaultTTL),
				ResourceRecords: []route53types.ResourceRecord{{Value: aws.String("8.8.8.8")}},
				SetIdentifier:   aws.String("geoproximity-region"),
				GeoProximityLocation: &route53types.GeoProximityLocation{
					AWSRegion: aws.String("us-west-2"),
					Bias:      aws.Int32(10),
				},
			},
			{
				Name:            aws.String("update-test-geoproximity.zone-1.ext-dns-test-2.teapot.zalan.do."),
				Type:            route53types.RRTypeA,
				TTL:             aws.Int64(defaultTTL),
				ResourceRecords: []route53types.ResourceRecord{{Value: aws.String("1.2.3.4")}},
				SetIdentifier:   aws.String("geoproximity-update"),
				GeoProximityLocation: &route53types.GeoProximityLocation{
					LocalZoneGroup: aws.String("usw2-phx2-az1"),
				},
			},
			{
				Name:            aws.String("create-test-geoproximity-coordinates.zone-1.ext-dns-test-2.teapot.zalan.do."),
				Type:            route53types.RRTypeA,
				TTL:             aws.Int64(defaultTTL),
				ResourceRecords: []route53types.ResourceRecord{{Value: aws.String("8.8.8.8")}},
				SetIdentifier:   aws.String("geoproximity-coordinates"),
				GeoProximityLocation: &route53types.GeoProximityLocation{
					Coordinates: &route53types.Coordinates{
						Latitude:  aws.String("60"),
						Longitude: aws.String("60"),
					},
				},
			},
		})
		validateRecords(t, listAWSRecords(t, provider.clients[defaultAWSProfile], "/hostedzone/zone-2.ext-dns-test-2.teapot.zalan.do."), []route53types.ResourceRecordSet{
			{
				Name:            aws.String("escape-\\045\\041s\\050\\074nil\\076\\051-codes.zone-2.ext-dns-test-2.teapot.zalan.do."),
				Type:            route53types.RRTypeA,
				TTL:             aws.Int64(defaultTTL),
				ResourceRecords: []route53types.ResourceRecord{{Value: aws.String("1.2.3.4")}},
				SetIdentifier:   aws.String("no-change"),
				Weight:          aws.Int64(10),
			},
			{
				Name:            aws.String("create-test.zone-2.ext-dns-test-2.teapot.zalan.do."),
				Type:            route53types.RRTypeA,
				TTL:             aws.Int64(defaultTTL),
				ResourceRecords: []route53types.ResourceRecord{{Value: aws.String("8.8.4.4")}},
			},
			{
				Name:            aws.String("create-test-aaaa.zone-2.ext-dns-test-2.teapot.zalan.do."),
				Type:            route53types.RRTypeAaaa,
				TTL:             aws.Int64(defaultTTL),
				ResourceRecords: []route53types.ResourceRecord{{Value: aws.String("2606:4700:4700::1001")}},
			},
			{
				Name:            aws.String("update-test.zone-2.ext-dns-test-2.teapot.zalan.do."),
				Type:            route53types.RRTypeA,
				TTL:             aws.Int64(defaultTTL),
				ResourceRecords: []route53types.ResourceRecord{{Value: aws.String("4.3.2.1")}},
			},
			{
				Name:            aws.String("update-test-aaaa.zone-2.ext-dns-test-2.teapot.zalan.do."),
				Type:            route53types.RRTypeAaaa,
				TTL:             aws.Int64(defaultTTL),
				ResourceRecords: []route53types.ResourceRecord{{Value: aws.String("2606:4700:4700::1111")}},
			},
			{
				Name:            aws.String("create-test-multiple.zone-2.ext-dns-test-2.teapot.zalan.do."),
				Type:            route53types.RRTypeA,
				TTL:             aws.Int64(defaultTTL),
				ResourceRecords: []route53types.ResourceRecord{{Value: aws.String("8.8.8.8")}, {Value: aws.String("8.8.4.4")}},
			},
			{
				Name:            aws.String("create-test-multiple-aaaa.zone-2.ext-dns-test-2.teapot.zalan.do."),
				Type:            route53types.RRTypeAaaa,
				TTL:             aws.Int64(defaultTTL),
				ResourceRecords: []route53types.ResourceRecord{{Value: aws.String("2606:4700:4700::1111")}, {Value: aws.String("2606:4700:4700::1001")}},
			},
			{
				Name:            aws.String("update-test-multiple.zone-2.ext-dns-test-2.teapot.zalan.do."),
				Type:            route53types.RRTypeA,
				TTL:             aws.Int64(defaultTTL),
				ResourceRecords: []route53types.ResourceRecord{{Value: aws.String("1.2.3.4")}, {Value: aws.String("4.3.2.1")}},
			},
			{
				Name:            aws.String("update-test-multiple-aaaa.zone-2.ext-dns-test-2.teapot.zalan.do."),
				Type:            route53types.RRTypeAaaa,
				TTL:             aws.Int64(defaultTTL),
				ResourceRecords: []route53types.ResourceRecord{{Value: aws.String("2606:4700:4700::1001")}, {Value: aws.String("2606:4700:4700::1111")}},
			},
			{
				Name:            aws.String("update-test-mx.zone-2.ext-dns-test-2.teapot.zalan.do."),
				Type:            route53types.RRTypeMx,
				TTL:             aws.Int64(defaultTTL),
				ResourceRecords: []route53types.ResourceRecord{{Value: aws.String("20 mailhost3.foo.elb.amazonaws.com")}},
			},
			{
				Name:            aws.String("update-test-naptr.zone-2.ext-dns-test-2.teapot.zalan.do."),
				Type:            route53types.RRTypeNaptr,
				TTL:             aws.Int64(defaultTTL),
				ResourceRecords: []route53types.ResourceRecord{{Value: aws.String(`20 "U" "SIP+DTU" "" _sip._udp.sip2.example.com.`)}},
			},
		})
	}
}

func TestAWSApplyChangesDryRun(t *testing.T) {
	originalRecords := []route53types.ResourceRecordSet{
		{
			Name:            aws.String("update-test.zone-1.ext-dns-test-2.teapot.zalan.do."),
			Type:            route53types.RRTypeA,
			TTL:             aws.Int64(defaultTTL),
			ResourceRecords: []route53types.ResourceRecord{{Value: aws.String("8.8.8.8")}},
		},
		{
			Name:            aws.String("delete-test.zone-1.ext-dns-test-2.teapot.zalan.do."),
			Type:            route53types.RRTypeA,
			TTL:             aws.Int64(defaultTTL),
			ResourceRecords: []route53types.ResourceRecord{{Value: aws.String("8.8.8.8")}},
		},
		{
			Name:            aws.String("update-test.zone-2.ext-dns-test-2.teapot.zalan.do."),
			Type:            route53types.RRTypeA,
			TTL:             aws.Int64(defaultTTL),
			ResourceRecords: []route53types.ResourceRecord{{Value: aws.String("8.8.4.4")}},
		},
		{
			Name:            aws.String("delete-test.zone-2.ext-dns-test-2.teapot.zalan.do."),
			Type:            route53types.RRTypeA,
			TTL:             aws.Int64(defaultTTL),
			ResourceRecords: []route53types.ResourceRecord{{Value: aws.String("8.8.4.4")}},
		},
		{
			Name:            aws.String("update-test-a-to-cname.zone-1.ext-dns-test-2.teapot.zalan.do."),
			Type:            route53types.RRTypeA,
			TTL:             aws.Int64(defaultTTL),
			ResourceRecords: []route53types.ResourceRecord{{Value: aws.String("1.1.1.1")}},
		},
		{
			Name:            aws.String("update-test-cname.zone-1.ext-dns-test-2.teapot.zalan.do."),
			Type:            route53types.RRTypeCname,
			TTL:             aws.Int64(defaultTTL),
			ResourceRecords: []route53types.ResourceRecord{{Value: aws.String("bar.elb.amazonaws.com")}},
		},
		{
			Name:            aws.String("delete-test-cname.zone-1.ext-dns-test-2.teapot.zalan.do."),
			Type:            route53types.RRTypeCname,
			TTL:             aws.Int64(defaultTTL),
			ResourceRecords: []route53types.ResourceRecord{{Value: aws.String("qux.elb.amazonaws.com")}},
		},
		{
			Name:            aws.String("update-test-cname-alias.zone-1.ext-dns-test-2.teapot.zalan.do."),
			Type:            route53types.RRTypeCname,
			TTL:             aws.Int64(defaultTTL),
			ResourceRecords: []route53types.ResourceRecord{{Value: aws.String("bar.elb.amazonaws.com")}},
		},
		{
			Name:            aws.String("delete-test-cname-alias.zone-1.ext-dns-test-2.teapot.zalan.do."),
			Type:            route53types.RRTypeCname,
			TTL:             aws.Int64(defaultTTL),
			ResourceRecords: []route53types.ResourceRecord{{Value: aws.String("qux.elb.amazonaws.com")}},
		},
		{
			Name:            aws.String("update-test-multiple.zone-2.ext-dns-test-2.teapot.zalan.do."),
			Type:            route53types.RRTypeA,
			TTL:             aws.Int64(defaultTTL),
			ResourceRecords: []route53types.ResourceRecord{{Value: aws.String("8.8.8.8")}, {Value: aws.String("8.8.4.4")}},
		},
		{
			Name:            aws.String("delete-test-multiple.zone-2.ext-dns-test-2.teapot.zalan.do."),
			Type:            route53types.RRTypeA,
			TTL:             aws.Int64(defaultTTL),
			ResourceRecords: []route53types.ResourceRecord{{Value: aws.String("1.2.3.4")}, {Value: aws.String("4.3.2.1")}},
		},
		{
			Name:            aws.String("update-test-mx.zone-1.ext-dns-test-2.teapot.zalan.do."),
			Type:            route53types.RRTypeMx,
			TTL:             aws.Int64(defaultTTL),
			ResourceRecords: []route53types.ResourceRecord{{Value: aws.String("20 mail.foo.elb.amazonaws.com")}},
		},
		{
			Name:            aws.String("delete-test-mx.zone-2.ext-dns-test-2.teapot.zalan.do."),
			Type:            route53types.RRTypeMx,
			TTL:             aws.Int64(defaultTTL),
			ResourceRecords: []route53types.ResourceRecord{{Value: aws.String("10 mail.bar.elb.amazonaws.com")}},
		},
	}

	provider, _ := newAWSProvider(t, endpoint.NewDomainFilter([]string{"ext-dns-test-2.teapot.zalan.do."}), provider.NewZoneIDFilter([]string{}), provider.NewZoneTypeFilter(""), defaultEvaluateTargetHealth, false, true, originalRecords)

	createRecords := []*endpoint.Endpoint{
		endpoint.NewEndpoint("create-test.zone-1.ext-dns-test-2.teapot.zalan.do", endpoint.RecordTypeA, "8.8.8.8"),
		endpoint.NewEndpoint("create-test.zone-2.ext-dns-test-2.teapot.zalan.do", endpoint.RecordTypeA, "8.8.4.4"),
		endpoint.NewEndpoint("create-test-cname.zone-1.ext-dns-test-2.teapot.zalan.do", endpoint.RecordTypeCNAME, "foo.elb.amazonaws.com"),
		endpoint.NewEndpoint("create-test-cname-alias.zone-1.ext-dns-test-2.teapot.zalan.do", endpoint.RecordTypeCNAME, "foo.elb.amazonaws.com"),
		endpoint.NewEndpoint("create-test-multiple.zone-2.ext-dns-test-2.teapot.zalan.do", endpoint.RecordTypeA, "8.8.8.8", "8.8.4.4"),
		endpoint.NewEndpoint("create-test-mx.zone-1.ext-dns-test-2.teapot.zalan.do", endpoint.RecordTypeMX, "30 mail.foo.elb.amazonaws.com"),
	}

	currentRecords := []*endpoint.Endpoint{
		endpoint.NewEndpoint("update-test.zone-1.ext-dns-test-2.teapot.zalan.do", endpoint.RecordTypeA, "8.8.8.8"),
		endpoint.NewEndpoint("update-test.zone-2.ext-dns-test-2.teapot.zalan.do", endpoint.RecordTypeA, "8.8.4.4"),
		endpoint.NewEndpoint("update-test-a-to-cname.zone-1.ext-dns-test-2.teapot.zalan.do", endpoint.RecordTypeA, "1.1.1.1"),
		endpoint.NewEndpoint("update-test-cname.zone-1.ext-dns-test-2.teapot.zalan.do", endpoint.RecordTypeCNAME, "bar.elb.amazonaws.com"),
		endpoint.NewEndpoint("update-test-cname-alias.zone-1.ext-dns-test-2.teapot.zalan.do", endpoint.RecordTypeCNAME, "bar.elb.amazonaws.com"),
		endpoint.NewEndpoint("update-test-multiple.zone-2.ext-dns-test-2.teapot.zalan.do", endpoint.RecordTypeA, "8.8.8.8", "8.8.4.4"),
		endpoint.NewEndpoint("update-test-mx.zone-1.ext-dns-test-2.teapot.zalan.do", endpoint.RecordTypeMX, "20 mail.foo.elb.amazonaws.com"),
	}
	updatedRecords := []*endpoint.Endpoint{
		endpoint.NewEndpoint("update-test.zone-1.ext-dns-test-2.teapot.zalan.do", endpoint.RecordTypeA, "1.2.3.4"),
		endpoint.NewEndpoint("update-test.zone-2.ext-dns-test-2.teapot.zalan.do", endpoint.RecordTypeA, "4.3.2.1"),
		endpoint.NewEndpoint("update-test-a-to-cname.zone-1.ext-dns-test-2.teapot.zalan.do", endpoint.RecordTypeCNAME, "foo.elb.amazonaws.com"),
		endpoint.NewEndpoint("update-test-cname.zone-1.ext-dns-test-2.teapot.zalan.do", endpoint.RecordTypeCNAME, "baz.elb.amazonaws.com"),
		endpoint.NewEndpoint("update-test-cname-alias.zone-1.ext-dns-test-2.teapot.zalan.do", endpoint.RecordTypeCNAME, "baz.elb.amazonaws.com"),
		endpoint.NewEndpoint("update-test-multiple.zone-2.ext-dns-test-2.teapot.zalan.do", endpoint.RecordTypeA, "1.2.3.4", "4.3.2.1"),
		endpoint.NewEndpoint("update-test-mx.zone-1.ext-dns-test-2.teapot.zalan.do", endpoint.RecordTypeMX, "10 mail.bar.elb.amazonaws.com"),
	}

	deleteRecords := []*endpoint.Endpoint{
		endpoint.NewEndpoint("delete-test.zone-1.ext-dns-test-2.teapot.zalan.do", endpoint.RecordTypeA, "8.8.8.8"),
		endpoint.NewEndpoint("delete-test.zone-2.ext-dns-test-2.teapot.zalan.do", endpoint.RecordTypeA, "8.8.4.4"),
		endpoint.NewEndpoint("delete-test-cname.zone-1.ext-dns-test-2.teapot.zalan.do", endpoint.RecordTypeCNAME, "qux.elb.amazonaws.com"),
		endpoint.NewEndpoint("delete-test-cname-alias.zone-1.ext-dns-test-2.teapot.zalan.do", endpoint.RecordTypeCNAME, "qux.elb.amazonaws.com"),
		endpoint.NewEndpoint("delete-test-multiple.zone-2.ext-dns-test-2.teapot.zalan.do", endpoint.RecordTypeA, "1.2.3.4", "4.3.2.1"),
		endpoint.NewEndpoint("delete-test-mx.zone-2.ext-dns-test-2.teapot.zalan.do", endpoint.RecordTypeMX, "10 mail.bar.elb.amazonaws.com"),
	}

	changes := &plan.Changes{
		Create:    createRecords,
		UpdateNew: updatedRecords,
		UpdateOld: currentRecords,
		Delete:    deleteRecords,
	}

	ctx := t.Context()

	require.NoError(t, provider.ApplyChanges(ctx, changes))

	validateRecords(t,
		append(
			listAWSRecords(t, provider.clients[defaultAWSProfile], "/hostedzone/zone-1.ext-dns-test-2.teapot.zalan.do."),
			listAWSRecords(t, provider.clients[defaultAWSProfile], "/hostedzone/zone-2.ext-dns-test-2.teapot.zalan.do.")...),
		originalRecords)
}
