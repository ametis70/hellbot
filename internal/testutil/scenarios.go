package testutil

// WarScenario returns a sequence of APIResponses that simulate a complete
// Helldivers 1 war from idle to war-won. The sequence is:
//
//  1. Idle — no events (first poll; poller saves but does not diff)
//  2. Attack event starts (active)
//  3. Attack event still active (no notification expected)
//  4. Attack event succeeded
//  5. Idle — no events
//  6. Defend event starts (active)
//  7. Defend event still active (no notification expected)
//  8. Defend event succeeded
//  9. War ends (season increments, all factions defeated → war won)
//
// Expected notifications (8 total):
//   - poll 2: attack started
//   - poll 4: attack succeeded
//   - poll 6: defend started
//   - poll 8: defend succeeded
//   - poll 9: war won
func WarScenario() []APIResponse {
	const (
		season   = 50
		t0       = 1_700_000_000
		attackID = 100
		defendID = 200
	)

	factions := func(status string) []APIFactionStatus {
		return []APIFactionStatus{
			{Season: season, Points: 300000, PointsMax: 300000, Status: status, IntroductionOrder: 2},
			{Season: season, Points: 400000, PointsMax: 400000, Status: status, IntroductionOrder: 1},
			{Season: season, Points: 200000, PointsMax: 200000, Status: status, IntroductionOrder: 0},
		}
	}

	factionsNewSeason := func() []APIFactionStatus {
		return []APIFactionStatus{
			{Season: season + 1, Points: 0, PointsMax: 300000, Status: "active", IntroductionOrder: 2},
			{Season: season + 1, Points: 0, PointsMax: 400000, Status: "active", IntroductionOrder: 1},
			{Season: season + 1, Points: 0, PointsMax: 200000, Status: "active", IntroductionOrder: 0},
		}
	}

	idle := func() APIResponse {
		return APIResponse{
			Time:          t0,
			FactionStatus: factions("active"),
			AttackEvents:  []APIAttackEvent{},
		}
	}

	return []APIResponse{
		// 1. First poll — idle, poller stores but does not diff
		idle(),

		// 2. Attack starts
		{
			Time:          t0 + 60,
			FactionStatus: factions("active"),
			AttackEvents: []APIAttackEvent{
				{Season: season, ID: attackID, StartTime: t0 + 60, EndTime: t0 + 3660, Enemy: 1, PointsMax: 50000, Status: "active", MaxEventID: attackID},
			},
		},

		// 3. Attack still active
		{
			Time:          t0 + 120,
			FactionStatus: factions("active"),
			AttackEvents: []APIAttackEvent{
				{Season: season, ID: attackID, StartTime: t0 + 60, EndTime: t0 + 3660, Enemy: 1, PointsMax: 50000, Points: 10000, Status: "active", MaxEventID: attackID},
			},
		},

		// 4. Attack succeeded
		{
			Time:          t0 + 180,
			FactionStatus: factions("active"),
			AttackEvents: []APIAttackEvent{
				{Season: season, ID: attackID, StartTime: t0 + 60, EndTime: t0 + 3660, Enemy: 1, PointsMax: 50000, Points: 50000, Status: "success", MaxEventID: attackID},
			},
		},

		// 5. Idle again
		idle(),

		// 6. Defend starts
		{
			Time:          t0 + 300,
			FactionStatus: factions("active"),
			DefendEvent: APIDefendEvent{
				Season: season, ID: defendID, StartTime: t0 + 300, EndTime: t0 + 4200, Region: 3, Enemy: 2, PointsMax: 30000, Status: "active",
			},
			AttackEvents: []APIAttackEvent{},
		},

		// 7. Defend still active
		{
			Time:          t0 + 360,
			FactionStatus: factions("active"),
			DefendEvent: APIDefendEvent{
				Season: season, ID: defendID, StartTime: t0 + 300, EndTime: t0 + 4200, Region: 3, Enemy: 2, PointsMax: 30000, Points: 5000, Status: "active",
			},
			AttackEvents: []APIAttackEvent{},
		},

		// 8. Defend succeeded, all factions now defeated (war ending next poll)
		{
			Time: t0 + 420,
			FactionStatus: []APIFactionStatus{
				{Season: season, Points: 300000, PointsMax: 300000, Status: "defeated", IntroductionOrder: 2},
				{Season: season, Points: 400000, PointsMax: 400000, Status: "defeated", IntroductionOrder: 1},
				{Season: season, Points: 200000, PointsMax: 200000, Status: "defeated", IntroductionOrder: 0},
			},
			DefendEvent: APIDefendEvent{
				Season: season, ID: defendID, StartTime: t0 + 300, EndTime: t0 + 4200, Region: 3, Enemy: 2, PointsMax: 30000, Points: 30000, Status: "success",
			},
			AttackEvents: []APIAttackEvent{},
		},

		// 9. War ends — season increments, previous factions all defeated → war won
		{
			Time:          t0 + 480,
			FactionStatus: factionsNewSeason(),
			AttackEvents:  []APIAttackEvent{},
		},
	}
}

// RestartScenario returns two separate sequences to simulate a bot restart
// mid-war. The bot runs on PartA (wars 50–52), is stopped, then restarted on
// PartB (starting at war 53). The first poll of PartB should be treated as a
// fresh start (no previous campaign in store) — the bot must not emit stale
// notifications for events that happened during the downtime.
//
// PartA sequence (war 50, bot stops after this):
//
//	poll 1: idle (stored, no diff)
//	poll 2: attack starts → notify started
//	poll 3: attack still active (no notification)
//
// PartB sequence (war 53, fresh store):
//
//	poll 1: war is already at season 53, attack 600 active → stored as baseline, no diff
//	poll 2: attack 600 still active → stored in event store, notify started
//	poll 3: attack 600 succeeded → notify succeeded
func RestartScenario() (partA, partB []APIResponse) {
	const t0 = 1_700_100_000

	partA = []APIResponse{
		// 1. Idle at season 50
		{
			Time: t0,
			FactionStatus: []APIFactionStatus{
				{Season: 50, Points: 300000, PointsMax: 300000, Status: "active"},
			},
			AttackEvents: []APIAttackEvent{},
		},
		// 2. Attack starts at season 50
		{
			Time: t0 + 60,
			FactionStatus: []APIFactionStatus{
				{Season: 50, Points: 300000, PointsMax: 300000, Status: "active"},
			},
			AttackEvents: []APIAttackEvent{
				{Season: 50, ID: 500, StartTime: t0 + 60, EndTime: t0 + 3660, Enemy: 1, PointsMax: 40000, Status: "active", MaxEventID: 500},
			},
		},
		// 3. Attack still active (bot stops here — state: attack 500 ongoing, season 50)
		{
			Time: t0 + 120,
			FactionStatus: []APIFactionStatus{
				{Season: 50, Points: 300000, PointsMax: 300000, Status: "active"},
			},
			AttackEvents: []APIAttackEvent{
				{Season: 50, ID: 500, StartTime: t0 + 60, EndTime: t0 + 3660, Enemy: 1, PointsMax: 40000, Points: 5000, Status: "active", MaxEventID: 500},
			},
		},
	}

	partB = []APIResponse{
		// 1. First poll after restart — season is now 53, attack 600 active.
		//    Store is empty: poller saves but does not diff (no previous campaign).
		{
			Time: t0 + 10000,
			FactionStatus: []APIFactionStatus{
				{Season: 53, Points: 100000, PointsMax: 300000, Status: "active"},
			},
			AttackEvents: []APIAttackEvent{
				{Season: 53, ID: 600, StartTime: t0 + 9000, EndTime: t0 + 13000, Enemy: 0, PointsMax: 60000, Points: 20000, Status: "active", MaxEventID: 600},
			},
		},
		// 2. Attack 600 still active — poller diffs, stores event 600, notifies started.
		{
			Time: t0 + 10060,
			FactionStatus: []APIFactionStatus{
				{Season: 53, Points: 100000, PointsMax: 300000, Status: "active"},
			},
			AttackEvents: []APIAttackEvent{
				{Season: 53, ID: 600, StartTime: t0 + 9000, EndTime: t0 + 13000, Enemy: 0, PointsMax: 60000, Points: 30000, Status: "active", MaxEventID: 600},
			},
		},
		// 3. Attack 600 succeeded — notify succeeded.
		{
			Time: t0 + 10120,
			FactionStatus: []APIFactionStatus{
				{Season: 53, Points: 100000, PointsMax: 300000, Status: "active"},
			},
			AttackEvents: []APIAttackEvent{
				{Season: 53, ID: 600, StartTime: t0 + 9000, EndTime: t0 + 13000, Enemy: 0, PointsMax: 60000, Points: 60000, Status: "success", MaxEventID: 600},
			},
		},
	}

	return partA, partB
}
