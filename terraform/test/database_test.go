/*
Copyright 2015-2021 Gravitational, Inc.

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

package test

import (
	"time"

	"github.com/gravitational/teleport/api/types"
	"github.com/gravitational/trace"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/stretchr/testify/require"
)

func (s *TerraformSuite) TestDatabase() {
	checkDestroyed := func(state *terraform.State) error {
		_, err := s.client.GetDatabase(s.Context(), "test")
		if trace.IsNotFound(err) {
			return nil
		}

		return err
	}

	name := "teleport_database.test"

	resource.Test(s.T(), resource.TestCase{
		ProtoV6ProviderFactories: s.terraformProviders,
		CheckDestroy:             checkDestroyed,
		Steps: []resource.TestStep{
			{
				Config: s.getFixture("database_0_create.tf"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(name, "kind", "db"),
					resource.TestCheckResourceAttr(name, "metadata.expires", "2032-10-12T07:20:50Z"),
					resource.TestCheckResourceAttr(name, "spec.protocol", "postgres"),
					resource.TestCheckResourceAttr(name, "spec.uri", "localhost"),
				),
			},
			{
				Config:   s.getFixture("database_0_create.tf"),
				PlanOnly: true,
			},
			{
				Config: s.getFixture("database_1_update.tf"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(name, "kind", "db"),
					resource.TestCheckResourceAttr(name, "metadata.expires", "2032-10-12T07:20:50Z"),
					resource.TestCheckResourceAttr(name, "spec.protocol", "postgres"),
					resource.TestCheckResourceAttr(name, "spec.uri", "example.com"),
				),
			},
			{
				Config:   s.getFixture("database_1_update.tf"),
				PlanOnly: true,
			},
		},
	})
}

func (s *TerraformSuite) TestImportDatabase() {
	r := "teleport_database"
	id := "test_import"
	name := r + "." + id

	database := &types.DatabaseV3{
		Metadata: types.Metadata{
			Name: id,
		},
		Spec: types.DatabaseSpecV3{
			Protocol: "postgres",
			URI:      "localhost:3000/test",
		},
	}
	err := database.CheckAndSetDefaults()
	require.NoError(s.T(), err)

	err = s.client.CreateDatabase(s.Context(), database)
	require.NoError(s.T(), err)

	require.Eventually(s.T(), func() bool {
		_, err := s.client.GetDatabase(s.Context(), database.GetName())
		if trace.IsNotFound(err) {
			return false
		}
		require.NoError(s.T(), err)
		return true
	}, 5*time.Second, time.Second)

	resource.Test(s.T(), resource.TestCase{
		ProtoV6ProviderFactories: s.terraformProviders,
		Steps: []resource.TestStep{
			{
				Config:        s.terraformConfig + "\n" + `resource "` + r + `" "` + id + `" { }`,
				ResourceName:  name,
				ImportState:   true,
				ImportStateId: id,
				ImportStateCheck: func(state []*terraform.InstanceState) error {
					require.Equal(s.T(), state[0].Attributes["kind"], "db")
					require.Equal(s.T(), state[0].Attributes["spec.uri"], "localhost:3000/test")
					require.Equal(s.T(), state[0].Attributes["spec.protocol"], "postgres")

					return nil
				},
			},
		},
	})
}
