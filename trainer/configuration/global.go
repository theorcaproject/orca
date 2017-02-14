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

type GlobalSettings struct {
	ApiPort                int
	LoggingPort            int
	CloudProvider          string
	AWSAccessKeyId         string
	AWSAccessKeySecret     string
	AWSRegion              string
	AWSBaseAmi             string
	AWSSSHKey              string
	AWSSSHKeyPath          string
	PlanningAlg            string
	InstanceUsername       string
	Uri                    string
	LoggingUri             string
	AWSSpotPrice           float64
	InstanceType           string
	SpotInstanceType       string

	AppChangeTimeout       int64
	ServerChangeTimeout    int64
	ServerTimeout          int64
	HostChangeFailureLimit int64

	Users                  map[string]User
	HostToken              string
}
