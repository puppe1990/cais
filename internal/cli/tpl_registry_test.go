package cli

import (
	"reflect"
	"testing"
)

// TestScaffoldTemplates_allRequiredPresent guards against accidental removal
// when editing tpl_scaffold_*.go files.
func TestScaffoldTemplates_allRequiredPresent(t *testing.T) {
	v := reflect.ValueOf(struct {
		TplGoMod              string
		TplEmptyCSS           string
		TplAir                string
		TplEnvExample         string
		TplWebEmbed           string
		TplMain               string
		TplMainBlank          string
		TplConsole            string
		TplApp                string
		TplAppBlank           string
		TplRoutes             string
		TplRoutesMinimal      string
		TplRoutesBlank        string
		TplHomeHandler        string
		TplHomeTest           string
		TplHomeTestMinimal    string
		TplContactHandler     string
		TplContactTest        string
		TplDashboardHandler   string
		TplDashboardTest      string
		TplHelpersTest        string
		TplGenericHandler     string
		TplGenericHandlerTest string
		TplContactModel       string
		TplStore              string
		TplStoreMinimal       string
		TplStoreTest          string
		TplStoreTestMinimal   string
		TplMigrations         string
		TplMigration001       string
		TplLayout             string
		TplLayoutWelcome      string
		TplLayoutMinimal      string
		TplLayoutBlank        string
		TplCaisLogo           string
		TplPageHome           string
		TplPageContact        string
		TplPageDashboard      string
		TplPartialErrors      string
		TplPartialSuccess     string
		TplGenericPage        string
		TplInputCSS           string
		TplTailwind           string
		TplPackageJSON        string
		TplMakefile           string
		TplCIWorkflow         string
		TplPreCommitConfig    string
		TplGolangci           string
		TplPrettierrc         string
		TplPrettierignore     string
		TplGitignore          string
		TplREADME             string
		TplREADMEBlank        string
		TplI18nCatalog        string
		TplI18nEn             string
		TplI18nPt             string
		TplI18nTest           string
		TplSeeds              string
		TplSeedsMinimal       string
		TplUserModel          string
		TplMigration002Auth   string
		TplAuthHandler        string
		TplStorePasswordReset string
		TplAuthTest           string
		TplPageLogin          string
		TplPageSignup         string
		TplPageForgotPassword string
		TplPageResetPassword  string
	}{
		TplGoMod:              tplGoMod,
		TplEmptyCSS:           tplEmptyCSS,
		TplAir:                tplAir,
		TplEnvExample:         tplEnvExample,
		TplWebEmbed:           tplWebEmbed,
		TplMain:               tplMain,
		TplMainBlank:          tplMainBlank,
		TplConsole:            tplConsole,
		TplApp:                tplApp,
		TplAppBlank:           tplAppBlank,
		TplRoutes:             tplRoutes,
		TplRoutesMinimal:      tplRoutesMinimal,
		TplRoutesBlank:        tplRoutesBlank,
		TplHomeHandler:        tplHomeHandler,
		TplHomeTest:           tplHomeTest,
		TplHomeTestMinimal:    tplHomeTestMinimal,
		TplContactHandler:     tplContactHandler,
		TplContactTest:        tplContactTest,
		TplDashboardHandler:   tplDashboardHandler,
		TplDashboardTest:      tplDashboardTest,
		TplHelpersTest:        tplHelpersTest,
		TplGenericHandler:     tplGenericHandler,
		TplGenericHandlerTest: tplGenericHandlerTest,
		TplContactModel:       tplContactModel,
		TplStore:              tplStore,
		TplStoreMinimal:       tplStoreMinimal,
		TplStoreTest:          tplStoreTest,
		TplStoreTestMinimal:   tplStoreTestMinimal,
		TplMigrations:         tplMigrations,
		TplMigration001:       tplMigration001,
		TplLayout:             tplLayout,
		TplLayoutWelcome:      tplLayoutWelcome,
		TplLayoutMinimal:      tplLayoutMinimal,
		TplLayoutBlank:        tplLayoutBlank,
		TplCaisLogo:           tplCaisLogo,
		TplPageHome:           tplPageHome,
		TplPageContact:        tplPageContact,
		TplPageDashboard:      tplPageDashboard,
		TplPartialErrors:      tplPartialErrors,
		TplPartialSuccess:     tplPartialSuccess,
		TplGenericPage:        tplGenericPage,
		TplInputCSS:           tplInputCSS,
		TplTailwind:           tplTailwind,
		TplPackageJSON:        tplPackageJSON,
		TplMakefile:           tplMakefile,
		TplCIWorkflow:         tplCIWorkflow,
		TplPreCommitConfig:    tplPreCommitConfig,
		TplGolangci:           tplGolangci,
		TplPrettierrc:         tplPrettierrc,
		TplPrettierignore:     tplPrettierignore,
		TplGitignore:          tplGitignore,
		TplREADME:             tplREADME,
		TplREADMEBlank:        tplREADMEBlank,
		TplI18nCatalog:        tplI18nCatalog,
		TplI18nEn:             tplI18nEn,
		TplI18nPt:             tplI18nPt,
		TplI18nTest:           tplI18nTest,
		TplSeeds:              tplSeeds,
		TplSeedsMinimal:       tplSeedsMinimal,
		TplUserModel:          tplUserModel,
		TplMigration002Auth:   tplMigration002Auth,
		TplAuthHandler:        tplAuthHandler,
		TplStorePasswordReset: tplStorePasswordReset,
		TplAuthTest:           tplAuthTest,
		TplPageLogin:          tplPageLogin,
		TplPageSignup:         tplPageSignup,
		TplPageForgotPassword: tplPageForgotPassword,
		TplPageResetPassword:  tplPageResetPassword,
	})

	for i := 0; i < v.NumField(); i++ {
		s := v.Field(i).String()
		if s == "" {
			t.Errorf("template %s is empty", v.Type().Field(i).Name)
		}
	}
}
