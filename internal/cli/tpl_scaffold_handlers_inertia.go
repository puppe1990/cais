// Inertia handler scaffold templates for cais new (handlers + tests).
// Split by domain so agents load one unit at a time:
//   tpl_scaffold_handlers_home.go / _contact.go / _dashboard.go / _auth.go
//   tpl_scaffold_handlers_hometest.go / _contacttest.go / _dashboardtest.go
//   tpl_scaffold_handlers_auth_tests.go / _testhelpers.go
// Note: do not use *_test.go for template const files (excluded from non-test builds).
package cli
