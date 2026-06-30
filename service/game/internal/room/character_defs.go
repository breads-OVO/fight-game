package room

// CharacterStats 角色基础属性
type CharacterStats struct {
	Name         string  // 角色名
	MaxHP        int32   // 最大生命值
	MoveSpeed    float64 // 移动速度
	JumpVelocity float64 // 跳跃速度
	LightDamage  int32   // 轻击伤害
	HeavyDamage  int32   // 重击伤害
	LightRange   float64 // 轻击范围
	HeavyRange   float64 // 重击范围
	LightStartup int32   // 轻击前摇帧
	HeavyStartup int32   // 重击前摇帧
	LightActive  int32   // 轻击有效帧
	HeavyActive  int32   // 重击有效帧
	LightHitstun int32   // 轻击命中硬直
	HeavyHitstun int32   // 重击命中硬直
	SkillDamage  int32   // 技能伤害
	SkillRange   float64 // 技能范围
	SkillStartup int32   // 技能前摇帧
	Weight       float64 // 重量（影响击退距离）
}

// SkillDef 技能定义
type SkillDef struct {
	Name         string  // 技能名
	Damage       int32   // 伤害
	Range        float64 // 范围
	Startup      int32   // 前摇帧
	Active       int32   // 有效帧
	Recovery     int32   // 后摇帧
	Hitstun      int32   // 命中硬直
	Knockback    float64 // 击退距离
	Special      int32   // 特性: 0=无, 1=升龙（无敌）, 2=突进, 3=飞行道具
	ProjectileVx float64 // 飞行道具速度（仅special=3）
	MeterCost    int32   // 气槽消耗
	CancelWindow int32   // 可取消的帧窗口（从第几帧开始可取消）
}

// characterDB 角色数据库
var characterDB = map[string]*CharacterStats{
	"1": {
		Name:         "疾风剑士",
		MaxHP:        100,
		MoveSpeed:    3.5,
		JumpVelocity: -10.5,
		LightDamage:  7,
		HeavyDamage:  18,
		LightRange:   75,
		HeavyRange:   95,
		LightStartup: 4,
		HeavyStartup: 10,
		LightActive:  3,
		HeavyActive:  4,
		LightHitstun: 10,
		HeavyHitstun: 18,
		SkillDamage:  25,
		SkillRange:   120,
		SkillStartup: 8,
		Weight:       1.0,
	},
	"2": {
		Name:         "铁壁卫士",
		MaxHP:        130,
		MoveSpeed:    2.2,
		JumpVelocity: -8.5,
		LightDamage:  6,
		HeavyDamage:  15,
		LightRange:   65,
		HeavyRange:   85,
		LightStartup: 6,
		HeavyStartup: 14,
		LightActive:  4,
		HeavyActive:  6,
		LightHitstun: 8,
		HeavyHitstun: 16,
		SkillDamage:  20,
		SkillRange:   100,
		SkillStartup: 10,
		Weight:       1.8,
	},
	"3": {
		Name:         "影之刺客",
		MaxHP:        80,
		MoveSpeed:    4.5,
		JumpVelocity: -12.0,
		LightDamage:  9,
		HeavyDamage:  22,
		LightRange:   70,
		HeavyRange:   90,
		LightStartup: 3,
		HeavyStartup: 9,
		LightActive:  2,
		HeavyActive:  3,
		LightHitstun: 12,
		HeavyHitstun: 20,
		SkillDamage:  30,
		SkillRange:   80,
		SkillStartup: 6,
		Weight:       0.7,
	},
	"4": {
		Name:         "炎术师",
		MaxHP:        90,
		MoveSpeed:    2.8,
		JumpVelocity: -9.0,
		LightDamage:  8,
		HeavyDamage:  20,
		LightRange:   80,
		HeavyRange:   100,
		LightStartup: 5,
		HeavyStartup: 12,
		LightActive:  3,
		HeavyActive:  5,
		LightHitstun: 10,
		HeavyHitstun: 18,
		SkillDamage:  28,
		SkillRange:   200,
		SkillStartup: 12,
		Weight:       0.9,
	},
	"5": {
		Name:         "雷霆拳手",
		MaxHP:        110,
		MoveSpeed:    3.8,
		JumpVelocity: -11.0,
		LightDamage:  10,
		HeavyDamage:  25,
		LightRange:   60,
		HeavyRange:   75,
		LightStartup: 2,
		HeavyStartup: 8,
		LightActive:  2,
		HeavyActive:  3,
		LightHitstun: 14,
		HeavyHitstun: 22,
		SkillDamage:  35,
		SkillRange:   90,
		SkillStartup: 5,
		Weight:       1.1,
	},
	"6": {
		Name:         "冰霜法师",
		MaxHP:        85,
		MoveSpeed:    3.0,
		JumpVelocity: -9.5,
		LightDamage:  7,
		HeavyDamage:  18,
		LightRange:   85,
		HeavyRange:   105,
		LightStartup: 6,
		HeavyStartup: 13,
		LightActive:  4,
		HeavyActive:  5,
		LightHitstun: 12,
		HeavyHitstun: 20,
		SkillDamage:  24,
		SkillRange:   180,
		SkillStartup: 14,
		Weight:       0.85,
	},
	"7": {
		Name:         "圣光骑士",
		MaxHP:        120,
		MoveSpeed:    3.0,
		JumpVelocity: -9.0,
		LightDamage:  7,
		HeavyDamage:  16,
		LightRange:   70,
		HeavyRange:   90,
		LightStartup: 5,
		HeavyStartup: 11,
		LightActive:  3,
		HeavyActive:  5,
		LightHitstun: 9,
		HeavyHitstun: 16,
		SkillDamage:  22,
		SkillRange:   110,
		SkillStartup: 9,
		Weight:       1.5,
	},
	"8": {
		Name:         "暗影射手",
		MaxHP:        90,
		MoveSpeed:    3.2,
		JumpVelocity: -10.0,
		LightDamage:  8,
		HeavyDamage:  20,
		LightRange:   90,
		HeavyRange:   150,
		LightStartup: 5,
		HeavyStartup: 15,
		LightActive:  3,
		HeavyActive:  4,
		LightHitstun: 10,
		HeavyHitstun: 18,
		SkillDamage:  26,
		SkillRange:   250,
		SkillStartup: 14,
		Weight:       0.8,
	},
	"9": {
		Name:         "狂战士",
		MaxHP:        140,
		MoveSpeed:    2.5,
		JumpVelocity: -8.0,
		LightDamage:  11,
		HeavyDamage:  28,
		LightRange:   65,
		HeavyRange:   80,
		LightStartup: 6,
		HeavyStartup: 15,
		LightActive:  4,
		HeavyActive:  6,
		LightHitstun: 10,
		HeavyHitstun: 18,
		SkillDamage:  40,
		SkillRange:   80,
		SkillStartup: 10,
		Weight:       2.0,
	},
	"10": {
		Name:         "风灵使者",
		MaxHP:        95,
		MoveSpeed:    4.2,
		JumpVelocity: -13.0,
		LightDamage:  6,
		HeavyDamage:  16,
		LightRange:   70,
		HeavyRange:   90,
		LightStartup: 3,
		HeavyStartup: 8,
		LightActive:  2,
		HeavyActive:  3,
		LightHitstun: 8,
		HeavyHitstun: 14,
		SkillDamage:  22,
		SkillRange:   70,
		SkillStartup: 4,
		Weight:       0.65,
	},
}

// skillDB 技能数据库（每个角色特有技能）
var skillDB = map[string]*SkillDef{
	"1":  {Name: "疾风斩", Damage: 25, Range: 120, Startup: 8, Active: 3, Recovery: 10, Hitstun: 20, Knockback: 40, Special: 2, ProjectileVx: 0, MeterCost: 1, CancelWindow: 6},
	"2":  {Name: "铁壁冲锋", Damage: 20, Range: 100, Startup: 10, Active: 4, Recovery: 14, Hitstun: 18, Knockback: 50, Special: 2, ProjectileVx: 0, MeterCost: 1, CancelWindow: 8},
	"3":  {Name: "影袭", Damage: 30, Range: 80, Startup: 6, Active: 2, Recovery: 8, Hitstun: 22, Knockback: 30, Special: 2, ProjectileVx: 0, MeterCost: 1, CancelWindow: 4},
	"4":  {Name: "火球术", Damage: 28, Range: 200, Startup: 12, Active: 1, Recovery: 16, Hitstun: 16, Knockback: 20, Special: 3, ProjectileVx: 5.0, MeterCost: 1, CancelWindow: 8},
	"5":  {Name: "雷霆一击", Damage: 35, Range: 90, Startup: 5, Active: 3, Recovery: 9, Hitstun: 24, Knockback: 60, Special: 2, ProjectileVx: 0, MeterCost: 1, CancelWindow: 3},
	"6":  {Name: "冰锥术", Damage: 24, Range: 180, Startup: 14, Active: 1, Recovery: 18, Hitstun: 18, Knockback: 15, Special: 3, ProjectileVx: 4.0, MeterCost: 1, CancelWindow: 10},
	"7":  {Name: "圣光突刺", Damage: 22, Range: 110, Startup: 9, Active: 3, Recovery: 12, Hitstun: 18, Knockback: 45, Special: 2, ProjectileVx: 0, MeterCost: 1, CancelWindow: 6},
	"8":  {Name: "穿心箭", Damage: 26, Range: 250, Startup: 14, Active: 1, Recovery: 16, Hitstun: 16, Knockback: 25, Special: 3, ProjectileVx: 6.0, MeterCost: 1, CancelWindow: 10},
	"9":  {Name: "狂暴重击", Damage: 40, Range: 80, Startup: 10, Active: 4, Recovery: 14, Hitstun: 26, Knockback: 70, Special: 1, ProjectileVx: 0, MeterCost: 1, CancelWindow: 7},
	"10": {Name: "风刃", Damage: 22, Range: 70, Startup: 4, Active: 2, Recovery: 7, Hitstun: 14, Knockback: 25, Special: 2, ProjectileVx: 0, MeterCost: 1, CancelWindow: 2},
}

// GetCharacterStats 获取角色属性
func GetCharacterStats(characterId string) *CharacterStats {
	if stats, ok := characterDB[characterId]; ok {
		return stats
	}
	// 默认值
	return &CharacterStats{
		Name:         "未知角色",
		MaxHP:        100,
		MoveSpeed:    3.0,
		JumpVelocity: -10.0,
		LightDamage:  8,
		HeavyDamage:  20,
		LightRange:   80,
		HeavyRange:   100,
		LightStartup: 5,
		HeavyStartup: 12,
		LightActive:  3,
		HeavyActive:  5,
		LightHitstun: 10,
		HeavyHitstun: 20,
		SkillDamage:  25,
		SkillRange:   100,
		SkillStartup: 8,
		Weight:       1.0,
	}
}

// GetCharacterNames 获取所有角色ID -> 名称映射
func GetCharacterNames() map[string]string {
	names := make(map[string]string, len(characterDB))
	for id, stats := range characterDB {
		names[id] = stats.Name
	}
	return names
}
