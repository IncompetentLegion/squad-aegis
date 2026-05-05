package logwatcher_manager

import (
	"context"
	"sync"
	"testing"
	"time"

	"github.com/google/uuid"
	"go.codycody31.dev/squad-aegis/internal/event_manager"
)

type testEventStore struct {
	mu           sync.Mutex
	serverID     uuid.UUID
	joinRequests map[string]*JoinRequestData
	playerData   map[string]*PlayerData
	sessionData  map[string]*SessionData
	roundWinner  *RoundWinnerData
	roundLoser   *RoundLoserData
	wonData      *WonData
}

func newTestEventStore(serverID uuid.UUID) *testEventStore {
	return &testEventStore{
		serverID:     serverID,
		joinRequests: make(map[string]*JoinRequestData),
		playerData:   make(map[string]*PlayerData),
		sessionData:  make(map[string]*SessionData),
	}
}

func (s *testEventStore) GetServerID() uuid.UUID { return s.serverID }

func (s *testEventStore) StoreJoinRequest(chainID string, playerData *JoinRequestData) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.joinRequests[chainID] = playerData
}

func (s *testEventStore) GetJoinRequest(chainID string) (*JoinRequestData, bool) {
	s.mu.Lock()
	defer s.mu.Unlock()
	value, ok := s.joinRequests[chainID]
	if ok {
		delete(s.joinRequests, chainID)
	}
	return value, ok
}

func (s *testEventStore) StorePlayerData(playerID string, data *PlayerData) {
	if playerID == "" {
		return
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	s.playerData[playerID] = data
}

func (s *testEventStore) GetPlayerData(playerID string) (*PlayerData, bool) {
	s.mu.Lock()
	defer s.mu.Unlock()
	value, ok := s.playerData[playerID]
	return value, ok
}

func (s *testEventStore) RemovePlayerData(playerID string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.playerData, playerID)
	return nil
}

func (s *testEventStore) StoreSessionData(key string, data *SessionData) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.sessionData[key] = data
}

func (s *testEventStore) GetSessionData(key string) (*SessionData, bool) {
	s.mu.Lock()
	defer s.mu.Unlock()
	value, ok := s.sessionData[key]
	return value, ok
}

func (s *testEventStore) StoreRoundWinner(data *RoundWinnerData) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.roundWinner = data
}

func (s *testEventStore) StoreRoundLoser(data *RoundLoserData) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.roundLoser = data
}

func (s *testEventStore) GetRoundWinner(remove bool) (*RoundWinnerData, bool) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.roundWinner == nil {
		return nil, false
	}
	value := s.roundWinner
	if remove {
		s.roundWinner = nil
	}
	return value, true
}

func (s *testEventStore) GetRoundLoser(remove bool) (*RoundLoserData, bool) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.roundLoser == nil {
		return nil, false
	}
	value := s.roundLoser
	if remove {
		s.roundLoser = nil
	}
	return value, true
}

func (s *testEventStore) StoreWonData(data *WonData) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.wonData = data
}

func (s *testEventStore) GetWonData() (*WonData, bool) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.wonData == nil {
		return nil, false
	}
	return s.wonData, true
}

func (s *testEventStore) ClearNewGameData() {}

func (s *testEventStore) CheckTeamkill(victimName string, attackerEOSID string) bool {
	return false
}

func (s *testEventStore) GetPlayerInfoByName(name string) (*event_manager.PlayerInfo, bool) {
	return nil, false
}

func (s *testEventStore) GetPlayerInfoByIdentifier(playerID string) (*event_manager.PlayerInfo, bool) {
	return nil, false
}

func (s *testEventStore) GetPlayerInfoByEOSID(eosID string) (*event_manager.PlayerInfo, bool) {
	return nil, false
}

func (s *testEventStore) GetPlayerInfoByController(controllerID string) (*event_manager.PlayerInfo, bool) {
	return nil, false
}

func waitForEvent(t *testing.T, ch <-chan event_manager.Event) event_manager.Event {
	t.Helper()

	select {
	case event := <-ch:
		return event
	case <-time.After(2 * time.Second):
		t.Fatal("timed out waiting for event")
		return event_manager.Event{}
	}
}

func TestProcessLogForEventsCarriesEpicAliasThroughJoinFlow(t *testing.T) {
	t.Parallel()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	em := event_manager.NewEventManager(ctx, 10)
	defer em.Shutdown()

	serverID := uuid.New()
	store := newTestEventStore(serverID)
	subscriber := em.Subscribe(event_manager.EventFilter{
		Types: []event_manager.EventType{
			event_manager.EventTypeLogPlayerConnected,
			event_manager.EventTypeLogJoinSucceeded,
		},
	}, nil, 10)
	defer em.Unsubscribe(subscriber.ID)

	parsers := GetLogParsers()

	postLogin := `[2026.03.29-12.00.00:000][1]LogSquad: PostLogin: NewPlayer: BP_PlayerController_C /Game/Maps/Narva.PersistentLevel.PlayerController_1 (IP: 127.0.0.1 | Online IDs: EOS: 0002bb228e4d4363ada0c139b11a9ece epic: e91067b2c8bb461ebf0cdf3a01ee5b0b)`
	joinSucceeded := `[2026.03.29-12.00.05:000][1]LogNet: Join succeeded: 13Baudouin`

	ProcessLogForEvents(postLogin, serverID, parsers, em, store, nil)
	connectedEvent := waitForEvent(t, subscriber.Channel)

	connectedData, ok := connectedEvent.Data.(*event_manager.LogPlayerConnectedData)
	if !ok {
		t.Fatalf("connected event data type = %T, want *LogPlayerConnectedData", connectedEvent.Data)
	}
	if connectedData.EOSID != "0002bb228e4d4363ada0c139b11a9ece" {
		t.Fatalf("connected EOSID = %q, want normalized EOS ID", connectedData.EOSID)
	}
	if connectedData.EpicID != "e91067b2c8bb461ebf0cdf3a01ee5b0b" {
		t.Fatalf("connected EpicID = %q, want normalized Epic ID", connectedData.EpicID)
	}

	ProcessLogForEvents(joinSucceeded, serverID, parsers, em, store, nil)
	joinEvent := waitForEvent(t, subscriber.Channel)

	joinData, ok := joinEvent.Data.(*event_manager.LogJoinSucceededData)
	if !ok {
		t.Fatalf("join event data type = %T, want *LogJoinSucceededData", joinEvent.Data)
	}
	if joinData.EOSID != "0002bb228e4d4363ada0c139b11a9ece" {
		t.Fatalf("join EOSID = %q, want normalized EOS ID", joinData.EOSID)
	}
	if joinData.EpicID != "e91067b2c8bb461ebf0cdf3a01ee5b0b" {
		t.Fatalf("join EpicID = %q, want normalized Epic ID", joinData.EpicID)
	}
}

func TestProcessLogForEventsParsesAdminToolsPostLogin(t *testing.T) {
	t.Parallel()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	em := event_manager.NewEventManager(ctx, 10)
	defer em.Shutdown()

	serverID := uuid.New()
	store := newTestEventStore(serverID)
	subscriber := em.Subscribe(event_manager.EventFilter{
		Types: []event_manager.EventType{event_manager.EventTypeLogPlayerConnected},
	}, nil, 10)
	defer em.Unsubscribe(subscriber.ID)

	parsers := GetLogParsers()

	postLogin := `[2026.05.05-08.38.13:583][806]LogSquad: PostLogin: NewPlayer: BP_PlayerController_AdminTools_C /SPM/Maps/Al_Basrah/Gameplay_Layer/SU_AlBasrah_AAS_v1.SU_AlBasrah_AAS_v1:PersistentLevel.BP_PlayerController_AdminTools_C_2147475584 (IP: 129.180.155.71 | Online IDs: EOS: 00020de8463c40df90cdef1395668ca1 steam: 76561198284214717)`

	ProcessLogForEvents(postLogin, serverID, parsers, em, store, nil)
	connectedEvent := waitForEvent(t, subscriber.Channel)

	connectedData, ok := connectedEvent.Data.(*event_manager.LogPlayerConnectedData)
	if !ok {
		t.Fatalf("connected event data type = %T, want *LogPlayerConnectedData", connectedEvent.Data)
	}
	if connectedData.PlayerController != "BP_PlayerController_AdminTools_C_2147475584" {
		t.Fatalf("connected PlayerController = %q, want AdminTools controller", connectedData.PlayerController)
	}
	if connectedData.IPAddress != "129.180.155.71" {
		t.Fatalf("connected IPAddress = %q, want 129.180.155.71", connectedData.IPAddress)
	}
	if connectedData.EOSID != "00020de8463c40df90cdef1395668ca1" {
		t.Fatalf("connected EOSID = %q, want normalized EOS ID", connectedData.EOSID)
	}
	if connectedData.SteamID != "76561198284214717" {
		t.Fatalf("connected SteamID = %q, want Steam ID", connectedData.SteamID)
	}
}

func TestProcessLogForEventsParsesEpicAliasOnPossess(t *testing.T) {
	t.Parallel()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	em := event_manager.NewEventManager(ctx, 10)
	defer em.Shutdown()

	serverID := uuid.New()
	store := newTestEventStore(serverID)
	subscriber := em.Subscribe(event_manager.EventFilter{
		Types: []event_manager.EventType{event_manager.EventTypeLogPlayerPossess},
	}, nil, 10)
	defer em.Unsubscribe(subscriber.ID)

	parsers := GetLogParsers()

	line := `[2026.03.29-12.01.00:000][2]LogSquadTrace: [DedicatedServer]ASQPlayerController::OnPossess(): PC=13Baudouin (Online IDs: EOS: 0002bb228e4d4363ada0c139b11a9ece epic: e91067b2c8bb461ebf0cdf3a01ee5b0b) Pawn=BP_Soldier_C`
	ProcessLogForEvents(line, serverID, parsers, em, store, nil)

	event := waitForEvent(t, subscriber.Channel)
	possessData, ok := event.Data.(*event_manager.LogPlayerPossessData)
	if !ok {
		t.Fatalf("possess event data type = %T, want *LogPlayerPossessData", event.Data)
	}
	if possessData.PlayerEOS != "0002bb228e4d4363ada0c139b11a9ece" {
		t.Fatalf("possess PlayerEOS = %q, want normalized EOS ID", possessData.PlayerEOS)
	}
	if possessData.PlayerEpic != "e91067b2c8bb461ebf0cdf3a01ee5b0b" {
		t.Fatalf("possess PlayerEpic = %q, want normalized Epic ID", possessData.PlayerEpic)
	}
}

func TestProcessLogForEventsParsesSuperModPossessClassname(t *testing.T) {
	t.Parallel()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	em := event_manager.NewEventManager(ctx, 10)
	defer em.Shutdown()

	serverID := uuid.New()
	store := newTestEventStore(serverID)
	subscriber := em.Subscribe(event_manager.EventFilter{
		Types: []event_manager.EventType{event_manager.EventTypeLogPlayerPossess},
	}, nil, 10)
	defer em.Unsubscribe(subscriber.ID)

	parsers := GetLogParsers()

	line := `[2026.05.05-08.44.12:421][ 71]LogSquadTrace: [DedicatedServer]OnPossess(): PC=ameer.mahmood175 (Online IDs: EOS: 0002871320b848a8ba45d79b47624d55 steam: 76561198776143392) Pawn=BP_AH-64D_2xM260_C_2147473926 FullPath=BP_AH-64D_2xM260_C /SPM/Maps/SEED/Gameplay_Layer/SU_Tallil_Seed_Vehicle_V1.SU_Tallil_Seed_Vehicle_V1:PersistentLevel.BP_AH-64D_2xM260_C_2147473926`
	ProcessLogForEvents(line, serverID, parsers, em, store, nil)

	event := waitForEvent(t, subscriber.Channel)
	possessData, ok := event.Data.(*event_manager.LogPlayerPossessData)
	if !ok {
		t.Fatalf("possess event data type = %T, want *LogPlayerPossessData", event.Data)
	}
	if possessData.PossessClassname != "BP_AH-64D_2xM260" {
		t.Fatalf("possess PossessClassname = %q, want SuperMod classname", possessData.PossessClassname)
	}
}

func TestProcessLogForEventsParsesAdminToolsDisconnect(t *testing.T) {
	t.Parallel()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	em := event_manager.NewEventManager(ctx, 10)
	defer em.Shutdown()

	serverID := uuid.New()
	store := newTestEventStore(serverID)
	subscriber := em.Subscribe(event_manager.EventFilter{
		Types: []event_manager.EventType{event_manager.EventTypeLogPlayerDisconnected},
	}, nil, 10)
	defer em.Unsubscribe(subscriber.ID)

	parsers := GetLogParsers()

	line := `[2026.05.05-08.42.13:993][989]LogNet: UChannel::Close: Sending CloseBunch. ChIndex == 0. Name: [UChannel] ChIndex: 0, Closing: 0 [UNetConnection] RemoteAddr: 79.177.155.14:3444, Name: RedpointEOSIpNetConnection_2147471505, Driver: Name:GameNetDriver Def:GameNetDriver RedpointEOSNetDriver_2147482314, IsServer: YES, PC: BP_PlayerController_AdminTools_C_2147470907, Owner: BP_PlayerController_AdminTools_C_2147470907, UniqueId: RedpointEOS:0002871320b848a8ba45d79b47624d55`
	ProcessLogForEvents(line, serverID, parsers, em, store, nil)

	event := waitForEvent(t, subscriber.Channel)
	disconnectData, ok := event.Data.(*event_manager.LogPlayerDisconnectedData)
	if !ok {
		t.Fatalf("disconnect event data type = %T, want *LogPlayerDisconnectedData", event.Data)
	}
	if disconnectData.PlayerController != "BP_PlayerController_AdminTools_C_2147470907" {
		t.Fatalf("disconnect PlayerController = %q, want AdminTools controller", disconnectData.PlayerController)
	}
	if disconnectData.IP != "79.177.155.14" {
		t.Fatalf("disconnect IP = %q, want 79.177.155.14", disconnectData.IP)
	}
	if disconnectData.EOSID != "0002871320b848a8ba45d79b47624d55" {
		t.Fatalf("disconnect EOSID = %q, want EOS ID", disconnectData.EOSID)
	}
}

func TestProcessLogForEventsParsesSuperModDeployableDamage(t *testing.T) {
	t.Parallel()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	em := event_manager.NewEventManager(ctx, 10)
	defer em.Shutdown()

	serverID := uuid.New()
	store := newTestEventStore(serverID)
	subscriber := em.Subscribe(event_manager.EventFilter{
		Types: []event_manager.EventType{event_manager.EventTypeLogDeployableDamaged},
	}, nil, 10)
	defer em.Unsubscribe(subscriber.ID)

	parsers := GetLogParsers()

	line := `[2026.05.05-17.57.49:209][908]LogSquadTrace: [DedicatedServer]TakeDamage(): BP_I_Sandbag_MurderHole_Desert_C_2145836233: 70.00 damage attempt by causer BP_ZSU-23-4_Weapon_INS_C_2145870221 instigator GlaDi with damage type BP_Kinetic_DamageType_C health remaining 225.00`
	ProcessLogForEvents(line, serverID, parsers, em, store, nil)

	event := waitForEvent(t, subscriber.Channel)
	damageData, ok := event.Data.(*event_manager.LogDeployableDamagedData)
	if !ok {
		t.Fatalf("deployable damage event data type = %T, want *LogDeployableDamagedData", event.Data)
	}
	if damageData.Deployable != "BP_I_Sandbag_MurderHole_Desert" {
		t.Fatalf("deployable damage Deployable = %q, want sandbag classname", damageData.Deployable)
	}
	if damageData.Weapon != "BP_ZSU-23-4_Weapon_INS" {
		t.Fatalf("deployable damage Weapon = %q, want SuperMod weapon classname", damageData.Weapon)
	}
	if damageData.DamageType != "BP_Kinetic_DamageType" {
		t.Fatalf("deployable damage DamageType = %q, want damage type without _C", damageData.DamageType)
	}
}

func TestProcessLogForEventsParsesSuperModCombatClassnames(t *testing.T) {
	t.Parallel()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	em := event_manager.NewEventManager(ctx, 10)
	defer em.Shutdown()

	serverID := uuid.New()
	store := newTestEventStore(serverID)
	subscriber := em.Subscribe(event_manager.EventFilter{
		Types: []event_manager.EventType{
			event_manager.EventTypeLogPlayerDamaged,
			event_manager.EventTypeLogPlayerWounded,
			event_manager.EventTypeLogPlayerDied,
		},
	}, nil, 10)
	defer em.Unsubscribe(subscriber.ID)

	parsers := GetLogParsers()

	damageLine := `[2026.05.05-17.03.52:242][580]LogSquad: Player:B3 n1mag1 ActualDamage=186.000015 from MNY Tyreza (Online IDs: EOS: 0002a0e299884b4084218a91a96ee3ab steam: 76561198315089650 | Player Controller ID: BP_PlayerController_AdminTools_C_2146128933)caused by BP_HK416F-S_SJ4_Foregrip_C_2146012323`
	ProcessLogForEvents(damageLine, serverID, parsers, em, store, nil)
	damageEvent := waitForEvent(t, subscriber.Channel)
	damageData, ok := damageEvent.Data.(*event_manager.LogPlayerDamagedData)
	if !ok {
		t.Fatalf("damage event data type = %T, want *LogPlayerDamagedData", damageEvent.Data)
	}
	if damageData.Weapon != "BP_HK416F-S_SJ4_Foregrip" {
		t.Fatalf("damage Weapon = %q, want SuperMod weapon classname", damageData.Weapon)
	}

	woundLine := `[2026.05.05-17.03.52:242][580]LogSquadTrace: [DedicatedServer]Wound(): Player:B3 n1mag1 KillingDamage=186.000015 from BP_PlayerController_AdminTools_C_2146128933 (Online IDs: EOS: 0002a0e299884b4084218a91a96ee3ab steam: 76561198315089650 | Controller ID: BP_PlayerController_AdminTools_C_2146128933) caused by BP_HK416F-S_SJ4_Foregrip_C_2146012323`
	ProcessLogForEvents(woundLine, serverID, parsers, em, store, nil)
	woundEvent := waitForEvent(t, subscriber.Channel)
	woundData, ok := woundEvent.Data.(*event_manager.LogPlayerWoundedData)
	if !ok {
		t.Fatalf("wound event data type = %T, want *LogPlayerWoundedData", woundEvent.Data)
	}
	if woundData.Weapon != "BP_HK416F-S_SJ4_Foregrip" {
		t.Fatalf("wound Weapon = %q, want SuperMod weapon classname", woundData.Weapon)
	}

	dieLine := `[2026.05.05-17.49.23:845][737]LogSquadTrace: [DedicatedServer]Die(): Player: [PACK] Sapper Rat KillingDamage=4.179260 from BP_PlayerController_AdminTools_C_2146105866 (Online IDs: EOS: 00027c6a42774a3398d8f569ac829ffc steam: 76561198196766079 | Contoller ID: BP_PlayerController_AdminTools_C_2146105866) caused by nullptr`
	ProcessLogForEvents(dieLine, serverID, parsers, em, store, nil)
	dieEvent := waitForEvent(t, subscriber.Channel)
	dieData, ok := dieEvent.Data.(*event_manager.LogPlayerDiedData)
	if !ok {
		t.Fatalf("die event data type = %T, want *LogPlayerDiedData", dieEvent.Data)
	}
	if dieData.Weapon != "nullptr" {
		t.Fatalf("die Weapon = %q, want nullptr", dieData.Weapon)
	}
	if dieData.AttackerSteam != "76561198196766079" {
		t.Fatalf("die AttackerSteam = %q, want Steam ID", dieData.AttackerSteam)
	}
}

func TestUnifiedGameParsersSkipAdminToolsBootstrapWorld(t *testing.T) {
	t.Parallel()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	em := event_manager.NewEventManager(ctx, 10)
	defer em.Shutdown()

	serverID := uuid.New()
	store := newTestEventStore(serverID)
	subscriber := em.Subscribe(event_manager.EventFilter{
		Types: []event_manager.EventType{event_manager.EventTypeLogGameEventUnified},
	}, nil, 10)
	defer em.Unsubscribe(subscriber.ID)

	parsers := GetUnifiedGameEventParsers()

	adminToolsWorld := `[2026.05.05-05.05.08:151][  0]LogWorld: Bringing World /SquadAdminTools/Maps/MapInitializer/Gameplay_Layers/VoiceConnect_Init.VoiceConnect_Init up for play (max tick rate 60) at 2026.05.05-05.05.08`
	ProcessLogForEvents(adminToolsWorld, serverID, parsers, em, store, nil)

	select {
	case event := <-subscriber.Channel:
		t.Fatalf("got unexpected event for admin tools bootstrap world: %#v", event.Data)
	case <-time.After(100 * time.Millisecond):
	}

	spmWorld := `[2026.05.05-16.27.37:851][626]LogWorld: Bringing World /SPM/Maps/Gorodok/Gameplay_Layer/PreCap/SPM_Gorodok_RAAS_v2_PreCap.SPM_Gorodok_RAAS_v2_PreCap up for play (max tick rate 60) at 2026.05.05-16.27.37`
	ProcessLogForEvents(spmWorld, serverID, parsers, em, store, nil)
	event := waitForEvent(t, subscriber.Channel)

	gameData, ok := event.Data.(*event_manager.LogGameEventUnifiedData)
	if !ok {
		t.Fatalf("game event data type = %T, want *LogGameEventUnifiedData", event.Data)
	}
	if gameData.DLC != "SPM" || gameData.MapClassname != "Gorodok" || gameData.LayerClassname != "SPM_Gorodok_RAAS_v2_PreCap" {
		t.Fatalf("game event map data = %q/%q/%q, want SPM/Gorodok/SPM_Gorodok_RAAS_v2_PreCap", gameData.DLC, gameData.MapClassname, gameData.LayerClassname)
	}
}
