// Copyright 2016 Netflix, Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package param

// properties
const (
	Enabled          = "chaosmonkey.enabled"
	Leashed          = "chaosmonkey.leashed"
	ScheduleEnabled  = "chaosmonkey.schedule_enabled"
	Accounts         = "chaosmonkey.accounts"
	StartHour        = "chaosmonkey.start_hour"
	EndHour          = "chaosmonkey.end_hour"
	TimeZone         = "chaosmonkey.time_zone"
	CronPath         = "chaosmonkey.cron_path"
	TermPath         = "chaosmonkey.term_path"
	TermAccount      = "chaosmonkey.term_account"
	MaxApps          = "chaosmonkey.max_apps"
	Trackers         = "chaosmonkey.trackers"
	ErrorCounter     = "chaosmonkey.error_counter"
	Decryptor        = "chaosmonkey.decryptor"
	OutageChecker    = "chaosmonkey.outage_checker"
	CronExpression   = "chaosmonkey.cron_expression"
	ScheduleCronPath = "chaosmonkey.schedule_cron_path"
	SchedulePath     = "chaosmonkey.schedule_path"
	LogPath          = "chaosmonkey.log_path"

	// spinnaker
	SpinnakerEndpoint          = "spinnaker.endpoint"
	SpinnakerCertificate       = "spinnaker.certificate"
	SpinnakerEncryptedPassword = "spinnaker.encrypted_password"
	SpinnakerUser              = "spinnaker.user"
	SpinnakerX509Cert          = "spinnaker.x509_cert"
	SpinnakerX509Key           = "spinnaker.x509_key"
	// database
	DatabaseHost              = "database.host"
	DatabasePort              = "database.port"
	DatabaseUser              = "database.user"
	DatabaseEncryptedPassword = "database.encrypted_password"
	DatabaseName              = "database.name"

	// dynamic property provider
	DynamicProvider = "dynamic.provider"
	DynamicEndpoint = "dynamic.endpoint"
	DynamicPath     = "dynamic.path"
)
