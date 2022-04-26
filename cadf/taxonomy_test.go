/*******************************************************************************
*
* Copyright 2022 SAP SE
*
* Licensed under the Apache License, Version 2.0 (the "License");
* you may not use this file except in compliance with the License.
* You should have received a copy of the License along with this
* program. If not, you may obtain a copy of the License at
*
*     http://www.apache.org/licenses/LICENSE-2.0
*
* Unless required by applicable law or agreed to in writing, software
* distributed under the License is distributed on an "AS IS" BASIS,
* WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
* See the License for the specific language governing permissions and
* limitations under the License.
*
*******************************************************************************/

package cadf

import "testing"

//nolint:dupl
func TestGetAction(t *testing.T) {
	type args struct {
		req string
	}
	tests := []struct {
		name       string
		args       args
		wantAction Action
	}{
		{"test1", args{req: "get"}, ReadAction},
		{"test2", args{req: "post"}, CreateAction},
		{"test3", args{req: "bork"}, UnknownAction},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if gotAction := GetAction(tt.args.req); gotAction != tt.wantAction {
				t.Errorf("GetAction() = %v, want %v", gotAction, tt.wantAction)
			}
		})
	}
}
