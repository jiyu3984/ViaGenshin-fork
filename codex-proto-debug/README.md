# GIO / VIA v6.5.0 Amber damage debug

This branch contains the current v6.5.0 proto subset that is directly relevant to the Amber arrow hit/no-damage issue.

Current observations:

- Normal attacks deal damage correctly.
- Amber arrows hit and produce `EvtBeingHitInfo`.
- v6.5 logs show `AttackResult.damage`, `damage_shield`, `attacker_id`, `defense_id`, `bullet_fly_time_ms`, `hit_collision`, `hit_eff_result`, and `ability_identifier`.
- Therefore `AttackResult.damage` tag is not the first suspect.
- The strongest suspects are projectile/bullet attacker lookup and the v6.5 -> v3.2 conversion path for arrow hits.
- Server log has shown: `[亡语伤害] attacker not exist, ignore damage`.

Check the same arrow by matching `attack_timestamp_ms` before and after conversion and compare:

- `attacker_id`
- `defense_id`
- `damage`
- `damage_shield`
- `ability_identifier.ability_caster_id`
- `ability_identifier.modifier_owner_id`
- `ability_identifier.instanced_ability_id`
- `ability_identifier.instanced_modifier_id`
- `ability_identifier.local_id`
- `bullet_fly_time_ms`
- `attack_count`
- `target_type`
- `element_type`

Also log the exact entity id passed to the server-side attacker lookup / `scene.getEntityById(...)` in projectile damage handling.

Priority:

1. projectile/bullet attacker lookup
2. v6.5 -> v3.2 `EvtBeingHitInfo` / `AttackResult` arrow conversion path
3. unresolved arrow-specific fields
4. other `AbilityIdentifier` subfields
5. `HitCollision` / `AttackHitEffectResult`
6. `EvtSetAttackTargetInfo`
