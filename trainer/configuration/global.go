/*
Copyright Alex Mack (al9mack@gmail.com) and Michael Lawson (michael@sphinix.com)
This file is part of Orca.

Orca is free software: you can redistribute it and/or modify
it under the terms of the GNU General Public License as published by
the Free Software Foundation, either version 3 of the License, or
(at your option) any later version.

Orca is distributed in the hope that it will be useful,
but WITHOUT ANY WARRANTY; without even the implied warranty of
MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
GNU General Public License for more details.

You should have received a copy of the GNU General Public License
along with Orca.  If not, see <http://www.gnu.org/licenses/>.
*/

package configuration

type User struct {
	Password string
}

type ApiToken struct {
	Token string
}

type AuditWebhook struct {
	Uri      string
	Severity string
}

type LoggingWebHook struct {
	Uri         string
	Certificate string
	User        string
	Password    string
}

type GlobalSettings struct {
	ApiPort                   int
	LoggingPort               int
	CloudProvider             string
	AWSAccessKeyId            string
	AWSAccessKeySecret        string
	AWSRegion                 string
	AWSBaseAmi                string
	AWSSSHKey                 string
	AWSSSHKeyPath             string
	PlanningAlg               string
	InstanceUsername          string
	Uri                       string
	LoggingUri                string
	AWSSpotPrice              float64
	InstanceType              string
	SpotInstanceType          string
	TrainerConfigBackupBucket string

	/* Immutable configuration for the boring planner.
	---> Much like AWS settings, must restart trainer for these
	---> to take effect
	*/
	AppChangeTimeout       int64
	ServerChangeTimeout    int64
	ServerTimeout          int64
	HostChangeFailureLimit int64
	ServerTTL              int64
	ServerCapacity         int64

	Users     map[string]User
	HostToken string

	ApiTokens       []ApiToken
	AuditWebhooks   []AuditWebhook
	LoggingWebHooks []LoggingWebHook

	AuditDatabaseUri string
	StatsDatabaseUri string

	EnvName          string
	PlanningDisabled bool
}
