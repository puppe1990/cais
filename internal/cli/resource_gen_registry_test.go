package cli

import (
	"reflect"
	"testing"
)

// TestResourceGenSymbols_present guards against accidental removal when
// editing resource_gen_*.go files.
func TestResourceGenSymbols_present(t *testing.T) {
	v := reflect.ValueOf(struct {
		ParseResourceOpts            any
		BuildResourceModel           any
		BuildReferenceStoreMethods   any
		BuildResourceMigration       any
		BuildResourceStoreMethods    any
		BuildResourcePaginatedMethod any
		BuildResourceSeed            any
		BuildResourceAdminHandler    any
		BuildResourcePublicHandler   any
		BuildAdminFormHTML           any
		BuildAdminIndexHTML          any
		BuildPublicListHTML          any
		TplSelectOptionModel         string
	}{
		ParseResourceOpts:            parseResourceOpts,
		BuildResourceModel:           buildResourceModel,
		BuildReferenceStoreMethods:   buildReferenceStoreMethods,
		BuildResourceMigration:       buildResourceMigration,
		BuildResourceStoreMethods:    buildResourceStoreMethods,
		BuildResourcePaginatedMethod: buildResourcePaginatedStoreMethod,
		BuildResourceSeed:            buildResourceSeed,
		BuildResourceAdminHandler:    buildResourceAdminHandler,
		BuildResourcePublicHandler:   buildResourcePublicHandler,
		BuildAdminFormHTML:           buildAdminFormHTML,
		BuildAdminIndexHTML:          buildAdminIndexHTML,
		BuildPublicListHTML:          buildPublicListHTML,
		TplSelectOptionModel:         tplSelectOptionModel,
	})

	for i := 0; i < v.NumField(); i++ {
		if v.Field(i).IsZero() {
			t.Errorf("missing resource gen symbol %s", v.Type().Field(i).Name)
		}
	}
}
