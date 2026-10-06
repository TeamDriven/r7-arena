// Copyright 2026 Team 254. All Rights Reserved.

package network

import (
	"fmt"
	"github.com/TeamDriven/r7-arena/model"
	"github.com/stretchr/testify/assert"
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"
)

func TestNewDnsMasq(t *testing.T) {
	dm := NewDnsMasq()
	assert.NotNil(t, dm)
	assert.Equal(t, "/etc/dnsmasq.d", dm.configDir)
	assert.NotNil(t, dm.restartService)
}

func TestDnsMasq_ConfigureTeamEthernet(t *testing.T) {
	tempDir := t.TempDir()
	restarted := false
	dm := &DnsMasq{
		configDir: tempDir,
		restartService: func() error {
			restarted = true
			return nil
		},
	}

	teams := [6]*model.Team{
		{Id: 254},
		{Id: 1114},
		nil,
		{Id: 2046},
		nil,
		nil,
	}

	assert.Nil(t, dm.ConfigureTeamEthernet(teams))
	assert.True(t, restarted)

	// Verify VLAN 10 (Red 1 - Team 254)
	vlan10Content, err := ioutil.ReadFile(filepath.Join(tempDir, "vlan10.conf"))
	assert.Nil(t, err)
	expectedVlan10 := "# Options for VLAN10\n" +
		"# Team 254\n\n" +
		"dhcp-range=set:vlan10,10.2.54.20,10.2.54.199,255.255.255.0,12h\n" +
		"dhcp-option=tag:vlan10,3,10.2.54.4\n"
	assert.Equal(t, expectedVlan10, string(vlan10Content))

	// Verify VLAN 20 (Red 2 - Team 1114)
	vlan20Content, err := ioutil.ReadFile(filepath.Join(tempDir, "vlan20.conf"))
	assert.Nil(t, err)
	expectedVlan20 := "# Options for VLAN20\n" +
		"# Team 1114\n\n" +
		"dhcp-range=set:vlan20,10.11.14.20,10.11.14.199,255.255.255.0,12h\n" +
		"dhcp-option=tag:vlan20,3,10.11.14.4\n"
	assert.Equal(t, expectedVlan20, string(vlan20Content))

	// Verify VLAN 40 (Blue 1 - Team 2046)
	vlan40Content, err := ioutil.ReadFile(filepath.Join(tempDir, "vlan40.conf"))
	assert.Nil(t, err)
	expectedVlan40 := "# Options for VLAN40\n" +
		"# Team 2046\n\n" +
		"dhcp-range=set:vlan40,10.20.46.20,10.20.46.199,255.255.255.0,12h\n" +
		"dhcp-option=tag:vlan40,3,10.20.46.4\n"
	assert.Equal(t, expectedVlan40, string(vlan40Content))

	// Reconfigure with updated teams: remove team 1114 and replace with team 1678
	restarted = false
	newTeams := [6]*model.Team{
		{Id: 254},
		{Id: 1678},
		nil,
		nil,
		nil,
		nil,
	}

	assert.Nil(t, dm.ConfigureTeamEthernet(newTeams))
	assert.True(t, restarted)

	// VLAN 10 should remain for Team 254
	assert.FileExists(t, filepath.Join(tempDir, "vlan10.conf"))

	// VLAN 20 should now contain Team 1678
	vlan20NewContent, err := ioutil.ReadFile(filepath.Join(tempDir, "vlan20.conf"))
	assert.Nil(t, err)
	assert.Contains(t, string(vlan20NewContent), "# Team 1678")
	assert.Contains(t, string(vlan20NewContent), "10.16.78.20")

	// VLAN 40 should be removed since no team is assigned
	_, err = os.Stat(filepath.Join(tempDir, "vlan40.conf"))
	assert.True(t, os.IsNotExist(err))
}

func TestDnsMasq_GetTeamVlans(t *testing.T) {
	tempDir := t.TempDir()
	dm := &DnsMasq{configDir: tempDir}

	// Create sample config files
	ioutil.WriteFile(filepath.Join(tempDir, "vlan10.conf"), []byte("# Options for VLAN10\n# Team 254\n"), 0644)
	ioutil.WriteFile(filepath.Join(tempDir, "vlan20.conf"), []byte("# Options for VLAN20\n# Team 1114\n"), 0644)
	ioutil.WriteFile(filepath.Join(tempDir, "vlan100.conf"), []byte("# Options for VLAN100\n# Team 9999\n"), 0644)
	ioutil.WriteFile(filepath.Join(tempDir, "other.txt"), []byte("some text"), 0644)

	teamVlans, err := dm.getTeamVlans()
	assert.Nil(t, err)
	assert.Equal(t, map[int]int{
		254:  10,
		1114: 20,
	}, teamVlans)

	// Error when directory does not exist
	invalidDm := &DnsMasq{configDir: filepath.Join(tempDir, "nonexistent")}
	_, err = invalidDm.getTeamVlans()
	assert.NotNil(t, err)
}

func TestDnsMasq_RestartError(t *testing.T) {
	tempDir := t.TempDir()
	dm := &DnsMasq{
		configDir: tempDir,
		restartService: func() error {
			return fmt.Errorf("systemctl failed")
		},
	}

	err := dm.ConfigureTeamEthernet([6]*model.Team{nil, nil, nil, nil, nil, nil})
	if assert.NotNil(t, err) {
		assert.Equal(t, "systemctl failed", err.Error())
	}
}
