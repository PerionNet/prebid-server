package ortb

import (
	"encoding/json"
	"testing"

	"github.com/prebid/openrtb/v17/openrtb2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/prebid/prebid-server/openrtb_ext"
	"github.com/prebid/prebid-server/util/ptrutil"
)

func TestSetDefaults(t *testing.T) {
	testCases := []struct {
		name            string
		givenRequest    openrtb2.BidRequest
		expectedRequest openrtb2.BidRequest
		expectedErr     string
	}{
		{
			name:            "empty",
			givenRequest:    openrtb2.BidRequest{},
			expectedRequest: openrtb2.BidRequest{},
		},
		{
			name:            "malformed request.ext",
			givenRequest:    openrtb2.BidRequest{Ext: json.RawMessage(`malformed`)},
			expectedRequest: openrtb2.BidRequest{Ext: json.RawMessage(`malformed`)},
			expectedErr:     "invalid character 'm' looking for beginning of value",
		},
		{
			name:            "targeting", // tests integration with setDefaultsTargeting
			givenRequest:    openrtb2.BidRequest{Ext: json.RawMessage(`{"prebid":{"targeting":{}}}`)},
			expectedRequest: openrtb2.BidRequest{Ext: json.RawMessage(`{"prebid":{"targeting":{"pricegranularity":{"precision":2,"ranges":[{"min":0,"max":20,"increment":0.1}]},"includewinners":true,"includebidderkeys":true}}}`)},
		},
	}

	for _, test := range testCases {
		t.Run(test.name, func(t *testing.T) {
			wrapper := &openrtb_ext.RequestWrapper{BidRequest: &test.givenRequest}

			// run
			err := SetDefaults(wrapper)

			// assert error
			if len(test.expectedErr) > 0 {
				assert.EqualError(t, err, test.expectedErr, "Error")
			}

			// rebuild request
			require.NoError(t, wrapper.RebuildRequest(), "Rebuild Request")

			// assert
			if len(test.expectedErr) > 0 {
				assert.EqualError(t, err, test.expectedErr, "Error")
				assert.Equal(t, &test.expectedRequest, wrapper.BidRequest, "Request")
			} else {
				// assert request as json to ignore order in ext fields
				expectedRequestJSON, err := json.Marshal(test.expectedRequest)
				require.NoError(t, err, "Marshal Expected Request")

				actualRequestJSON, err := json.Marshal(wrapper.BidRequest)
				require.NoError(t, err, "Marshal Actual Request")

				assert.JSONEq(t, string(expectedRequestJSON), string(actualRequestJSON), "Request")
			}
		})
	}
}

func TestSetDefaultsTargeting(t *testing.T) {
	defaultGranularity := openrtb_ext.PriceGranularity{
		Precision: ptrutil.ToPtr(DefaultPriceGranularityPrecision),
		Ranges:    []openrtb_ext.GranularityRange{{Min: 0, Max: 20, Increment: 0.1}},
	}

	testCases := []struct {
		name              string
		givenTargeting    *openrtb_ext.ExtRequestTargeting
		expectedTargeting *openrtb_ext.ExtRequestTargeting
		expectedModified  bool
	}{
		{
			name:              "nil",
			givenTargeting:    nil,
			expectedTargeting: nil,
		},
		{
			name:           "empty-targeting",
			givenTargeting: &openrtb_ext.ExtRequestTargeting{},
			expectedTargeting: &openrtb_ext.ExtRequestTargeting{
				PriceGranularity:  &defaultGranularity,
				IncludeWinners:    ptrutil.ToPtr(DefaultTargetingIncludeWinners),
				IncludeBidderKeys: ptrutil.ToPtr(DefaultTargetingIncludeBidderKeys),
			},
			expectedModified: true,
		},
		{
			name: "populated-partial", // precision and includewinners defaults set
			givenTargeting: &openrtb_ext.ExtRequestTargeting{
				PriceGranularity: &openrtb_ext.PriceGranularity{
					Ranges: []openrtb_ext.GranularityRange{{Min: 0, Max: 10, Increment: 1}},
				},
				IncludeBidderKeys: ptrutil.ToPtr(false),
			},
			expectedTargeting: &openrtb_ext.ExtRequestTargeting{
				PriceGranularity: &openrtb_ext.PriceGranularity{
					Precision: ptrutil.ToPtr(DefaultPriceGranularityPrecision),
					Ranges:    []openrtb_ext.GranularityRange{{Min: 0, Max: 10, Increment: 1}},
				},
				IncludeWinners:    ptrutil.ToPtr(DefaultTargetingIncludeWinners),
				IncludeBidderKeys: ptrutil.ToPtr(false),
			},
			expectedModified: true,
		},
		{
			name: "populated-no-granularity",
			givenTargeting: &openrtb_ext.ExtRequestTargeting{
				PriceGranularity:  &openrtb_ext.PriceGranularity{},
				IncludeWinners:    ptrutil.ToPtr(false),
				IncludeBidderKeys: ptrutil.ToPtr(false),
			},
			expectedTargeting: &openrtb_ext.ExtRequestTargeting{
				PriceGranularity:  &defaultGranularity,
				IncludeWinners:    ptrutil.ToPtr(false),
				IncludeBidderKeys: ptrutil.ToPtr(false),
			},
			expectedModified: true,
		},
		{
			name: "populated-ranges-nil",
			givenTargeting: &openrtb_ext.ExtRequestTargeting{
				PriceGranularity: &openrtb_ext.PriceGranularity{
					Precision: ptrutil.ToPtr(4),
					Ranges:    nil,
				},
			},
			expectedTargeting: &openrtb_ext.ExtRequestTargeting{
				PriceGranularity:  &defaultGranularity,
				IncludeWinners:    ptrutil.ToPtr(DefaultTargetingIncludeWinners),
				IncludeBidderKeys: ptrutil.ToPtr(DefaultTargetingIncludeBidderKeys),
			},
			expectedModified: true,
		},
		{
			name: "populated-ranges-empty",
			givenTargeting: &openrtb_ext.ExtRequestTargeting{
				PriceGranularity: &openrtb_ext.PriceGranularity{
					Precision: ptrutil.ToPtr(4),
					Ranges:    []openrtb_ext.GranularityRange{},
				},
			},
			expectedTargeting: &openrtb_ext.ExtRequestTargeting{
				PriceGranularity:  &defaultGranularity,
				IncludeWinners:    ptrutil.ToPtr(DefaultTargetingIncludeWinners),
				IncludeBidderKeys: ptrutil.ToPtr(DefaultTargetingIncludeBidderKeys),
			},
			expectedModified: true,
		},
		{
			name: "populated-full", // no defaults set
			givenTargeting: &openrtb_ext.ExtRequestTargeting{
				PriceGranularity: &openrtb_ext.PriceGranularity{
					Precision: ptrutil.ToPtr(4),
					Ranges:    []openrtb_ext.GranularityRange{{Min: 0, Max: 10, Increment: 1}},
				},
				IncludeWinners:    ptrutil.ToPtr(false),
				IncludeBidderKeys: ptrutil.ToPtr(false),
			},
			expectedTargeting: &openrtb_ext.ExtRequestTargeting{
				PriceGranularity: &openrtb_ext.PriceGranularity{
					Precision: ptrutil.ToPtr(4),
					Ranges:    []openrtb_ext.GranularityRange{{Min: 0, Max: 10, Increment: 1}},
				},
				IncludeWinners:    ptrutil.ToPtr(false),
				IncludeBidderKeys: ptrutil.ToPtr(false),
			},
			expectedModified: false,
		},
		{
			name: "setDefaultsPriceGranularity-integration",
			givenTargeting: &openrtb_ext.ExtRequestTargeting{
				PriceGranularity: &openrtb_ext.PriceGranularity{
					Precision: ptrutil.ToPtr(4),
					Ranges:    []openrtb_ext.GranularityRange{{Min: 5, Max: 10, Increment: 1}},
				},
				IncludeWinners:    ptrutil.ToPtr(false),
				IncludeBidderKeys: ptrutil.ToPtr(false),
			},
			expectedTargeting: &openrtb_ext.ExtRequestTargeting{
				PriceGranularity: &openrtb_ext.PriceGranularity{
					Precision: ptrutil.ToPtr(4),
					Ranges:    []openrtb_ext.GranularityRange{{Min: 0, Max: 10, Increment: 1}},
				},
				IncludeWinners:    ptrutil.ToPtr(false),
				IncludeBidderKeys: ptrutil.ToPtr(false),
			},
			expectedModified: true,
		},
	}

	for _, test := range testCases {
		t.Run(test.name, func(t *testing.T) {
			actualModified := setDefaultsTargeting(test.givenTargeting)
			assert.Equal(t, test.expectedModified, actualModified)
			assert.Equal(t, test.expectedTargeting, test.givenTargeting)
		})
	}
}

func TestSetDefaultsPriceGranularity(t *testing.T) {
	testCases := []struct {
		name                string
		givenGranularity    *openrtb_ext.PriceGranularity
		expectedGranularity *openrtb_ext.PriceGranularity
		expectedModified    bool
	}{
		{
			name: "no-precision",
			givenGranularity: &openrtb_ext.PriceGranularity{
				Precision: nil,
				Ranges:    []openrtb_ext.GranularityRange{{Min: 0, Max: 10, Increment: 1}},
			},
			expectedGranularity: &openrtb_ext.PriceGranularity{
				Precision: ptrutil.ToPtr(2),
				Ranges:    []openrtb_ext.GranularityRange{{Min: 0, Max: 10, Increment: 1}},
			},
			expectedModified: true,
		},
		{
			name: "incomplete-range",
			givenGranularity: &openrtb_ext.PriceGranularity{
				Precision: ptrutil.ToPtr(2),
				Ranges:    []openrtb_ext.GranularityRange{{Min: 5, Max: 10, Increment: 1}},
			},
			expectedGranularity: &openrtb_ext.PriceGranularity{
				Precision: ptrutil.ToPtr(2),
				Ranges:    []openrtb_ext.GranularityRange{{Min: 0, Max: 10, Increment: 1}},
			},
			expectedModified: true,
		},
		{
			name: "no-precision+incomplete-range",
			givenGranularity: &openrtb_ext.PriceGranularity{
				Precision: nil,
				Ranges:    []openrtb_ext.GranularityRange{{Min: 5, Max: 10, Increment: 1}},
			},
			expectedGranularity: &openrtb_ext.PriceGranularity{
				Precision: ptrutil.ToPtr(2),
				Ranges:    []openrtb_ext.GranularityRange{{Min: 0, Max: 10, Increment: 1}},
			},
			expectedModified: true,
		},
		{
			name: "all-set",
			givenGranularity: &openrtb_ext.PriceGranularity{
				Precision: ptrutil.ToPtr(2),
				Ranges:    []openrtb_ext.GranularityRange{{Min: 0, Max: 10, Increment: 1}},
			},
			expectedGranularity: &openrtb_ext.PriceGranularity{
				Precision: ptrutil.ToPtr(2),
				Ranges:    []openrtb_ext.GranularityRange{{Min: 0, Max: 10, Increment: 1}},
			},
			expectedModified: false,
		},
	}

	for _, test := range testCases {
		t.Run(test.name, func(t *testing.T) {
			actualModified := setDefaultsPriceGranularity(test.givenGranularity)
			assert.Equal(t, test.expectedModified, actualModified)
			assert.Equal(t, test.expectedGranularity, test.givenGranularity)
		})
	}
}

func TestSetDefaultsPriceGranularityRange(t *testing.T) {
	testCases := []struct {
		name             string
		givenRange       []openrtb_ext.GranularityRange
		expectedRange    []openrtb_ext.GranularityRange
		expectedModified bool
	}{
		{
			name:             "nil",
			givenRange:       nil,
			expectedRange:    nil,
			expectedModified: false,
		},
		{
			name:             "empty",
			givenRange:       []openrtb_ext.GranularityRange{},
			expectedRange:    []openrtb_ext.GranularityRange{},
			expectedModified: false,
		},
		{
			name:             "one-ok",
			givenRange:       []openrtb_ext.GranularityRange{{Min: 0, Max: 10, Increment: 1}},
			expectedRange:    []openrtb_ext.GranularityRange{{Min: 0, Max: 10, Increment: 1}},
			expectedModified: false,
		},
		{
			name:             "one-fixed",
			givenRange:       []openrtb_ext.GranularityRange{{Min: 5, Max: 10, Increment: 1}},
			expectedRange:    []openrtb_ext.GranularityRange{{Min: 0, Max: 10, Increment: 1}},
			expectedModified: true,
		},
		{
			name:             "many-ok",
			givenRange:       []openrtb_ext.GranularityRange{{Min: 0, Max: 10, Increment: 1}, {Min: 10, Max: 20, Increment: 1}},
			expectedRange:    []openrtb_ext.GranularityRange{{Min: 0, Max: 10, Increment: 1}, {Min: 10, Max: 20, Increment: 1}},
			expectedModified: false,
		},
		{
			name:             "many-fixed",
			givenRange:       []openrtb_ext.GranularityRange{{Min: 0, Max: 10, Increment: 1}, {Min: 15, Max: 20, Increment: 1}},
			expectedRange:    []openrtb_ext.GranularityRange{{Min: 0, Max: 10, Increment: 1}, {Min: 10, Max: 20, Increment: 1}},
			expectedModified: true,
		},
	}

	for _, test := range testCases {
		t.Run(test.name, func(t *testing.T) {
			actualModified := setDefaultsPriceGranularityRange(test.givenRange)
			assert.Equal(t, test.expectedModified, actualModified)
			assert.Equal(t, test.expectedRange, test.givenRange)
		})
	}
}
