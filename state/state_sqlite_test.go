package state

import (
	"io/ioutil"
	"testing"

	"github.com/status-im/mvds/persistenceutil"
	"github.com/status-im/mvds/state/migrations"
	"github.com/stretchr/testify/require"
)

func TestPersistentSyncState(t *testing.T) {
	tmpFile, err := ioutil.TempFile("", "")
	require.NoError(t, err)
	db, err := persistenceutil.Open(tmpFile.Name(), "", persistenceutil.MigrationConfig{
		AssetNames:  migrations.AssetNames(),
		AssetGetter: migrations.Asset,
	})
	require.NoError(t, err)
	p := NewPersistentSyncState(db)

	stateWithoutGroupID := State{
		Type:      OFFER,
		SendCount: 1,
		SendEpoch: 1,
		GroupID:   nil,
		PeerID:    PeerID{0x01},
		MessageID: MessageID{0xaa},
	}
	err = p.Add(stateWithoutGroupID)
	require.NoError(t, err)

	stateWithGroupID := stateWithoutGroupID
	stateWithGroupID.GroupID = &GroupID{0x01}
	stateWithGroupID.MessageID = MessageID{0xbb}
	err = p.Add(stateWithGroupID)
	require.NoError(t, err)

	// Getting states for the old epoch.
	allStates, err := p.All(0)
	require.NoError(t, err)
	require.Nil(t, allStates)

	// Getting states for the current epoch.
	allStates, err = p.All(1)
	require.NoError(t, err)
	require.Equal(t, []State{stateWithoutGroupID, stateWithGroupID}, allStates)
	require.Nil(t, allStates[0].GroupID)
	require.EqualValues(t, &GroupID{0x01}, allStates[1].GroupID)

	err = p.Remove(stateWithoutGroupID.MessageID, stateWithoutGroupID.PeerID)
	require.NoError(t, err)
	// remove non-existing row
	err = p.Remove(MessageID{0xff}, PeerID{0xff})
	require.EqualError(t, err, "state not found")
}

func TestMapStateWithPeerID(t *testing.T) {
	tmpFile, err := ioutil.TempFile("", "")
	require.NoError(t, err)
	db, err := persistenceutil.Open(tmpFile.Name(), "", persistenceutil.MigrationConfig{
		AssetNames:  migrations.AssetNames(),
		AssetGetter: migrations.Asset,
	})
	require.NoError(t, err)
	p := NewPersistentSyncState(db)

	stateForOnePeer := State{
		Type:      MESSAGE,
		SendCount: 1,
		SendEpoch: 1,
		GroupID:   &GroupID{0x01},
		PeerID:    PeerID{0x01},
		MessageID: MessageID{0xaa},
	}
	err = p.Add(stateForOnePeer)
	require.NoError(t, err)

	stateWithAnotherPeer := stateForOnePeer
	stateWithAnotherPeer.PeerID = PeerID{0x02}
	stateWithAnotherPeer.MessageID = MessageID{0xbb}
	err = p.Add(stateWithAnotherPeer)
	require.NoError(t, err)

	state2ForOnePeer := stateForOnePeer
	state2ForOnePeer.MessageID = MessageID{0xcc}
	err = p.Add(state2ForOnePeer)
	require.NoError(t, err)

	peerStates, err := p.QueryByPeerID(PeerID{0x01}, 3)
	require.NoError(t, err)
	require.Equal(t, []State{stateForOnePeer, state2ForOnePeer}, peerStates)

	peerStates, err = p.QueryByPeerID(PeerID{0x01}, 1)
	require.NoError(t, err)
	require.Equal(t, []State{stateForOnePeer}, peerStates)

	p.MapWithPeerId(PeerID{0x01}, func(state State) State {
		state.SendEpoch++
		return state
	})
	newState, err := p.QueryByPeerID(PeerID{0x01}, 3)
	require.NoError(t, err)
	expectedNewState := stateForOnePeer
	expectedNewState.SendEpoch++
	expectedNewState2 := state2ForOnePeer
	expectedNewState2.SendEpoch++
	require.Equal(t, []State{expectedNewState, expectedNewState2}, newState)
}
