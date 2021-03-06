package model

import (
	"encoding/json"
	es_models "github.com/caos/zitadel/internal/eventstore/models"
	"testing"
)

func TestAddressChanges(t *testing.T) {
	type args struct {
		existing *Address
		new      *Address
	}
	type res struct {
		changesLen int
	}
	tests := []struct {
		name string
		args args
		res  res
	}{
		{
			name: "all fields changed",
			args: args{
				existing: &Address{Country: "Country", Locality: "Locality", PostalCode: "PostalCode", Region: "Region", StreetAddress: "StreetAddress"},
				new:      &Address{Country: "CountryChanged", Locality: "LocalityChanged", PostalCode: "PostalCodeChanged", Region: "RegionChanged", StreetAddress: "StreetAddressChanged"},
			},
			res: res{
				changesLen: 5,
			},
		},
		{
			name: "no fields changed",
			args: args{
				existing: &Address{Country: "Country", Locality: "Locality", PostalCode: "PostalCode", Region: "Region", StreetAddress: "StreetAddress"},
				new:      &Address{Country: "Country", Locality: "Locality", PostalCode: "PostalCode", Region: "Region", StreetAddress: "StreetAddress"},
			},
			res: res{
				changesLen: 0,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			changes := tt.args.existing.Changes(tt.args.new)
			if len(changes) != tt.res.changesLen {
				t.Errorf("got wrong changes len: expected: %v, actual: %v ", tt.res.changesLen, len(changes))
			}
		})
	}
}

func TestAppendUserAddressChangedEvent(t *testing.T) {
	type args struct {
		user    *User
		address *Address
		event   *es_models.Event
	}
	tests := []struct {
		name   string
		args   args
		result *User
	}{
		{
			name: "append user address event",
			args: args{
				user:    &User{Address: &Address{Locality: "Locality", Country: "Country"}},
				address: &Address{Locality: "LocalityChanged", PostalCode: "PostalCode"},
				event:   &es_models.Event{},
			},
			result: &User{Address: &Address{Locality: "LocalityChanged", Country: "Country", PostalCode: "PostalCode"}},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.args.address != nil {
				data, _ := json.Marshal(tt.args.address)
				tt.args.event.Data = data
			}
			tt.args.user.appendUserAddressChangedEvent(tt.args.event)
			if tt.args.user.Address.Locality != tt.result.Address.Locality {
				t.Errorf("got wrong result: expected: %v, actual: %v ", tt.result, tt.args.user)
			}
			if tt.args.user.Address.Country != tt.result.Address.Country {
				t.Errorf("got wrong result: expected: %v, actual: %v ", tt.result, tt.args.user)
			}
			if tt.args.user.Address.PostalCode != tt.result.Address.PostalCode {
				t.Errorf("got wrong result: expected: %v, actual: %v ", tt.result, tt.args.user)
			}
		})
	}
}
