package database

import "time"

type AdminUser struct {
	ID           int64       `gorm:"column:id;primaryKey;autoIncrement"`
	Username     string      `gorm:"column:username;type:varchar(50);not null;uniqueIndex"`
	PasswordHash string      `gorm:"column:password_hash;type:varchar(255);not null"`
	FullName     string      `gorm:"column:full_name;type:varchar(100);not null"`
	Email        *string     `gorm:"column:email;type:varchar(100);uniqueIndex"`
	PhoneNumber  *string     `gorm:"column:phone_number;type:varchar(20)"`
	IsSuperAdmin bool        `gorm:"column:is_super_admin;not null;default:false;index"`
	IsActive     bool        `gorm:"column:is_active;not null;default:true;index"`
	CreatedAt    time.Time   `gorm:"column:created_at;not null;default:now()"`
	UpdatedAt    time.Time   `gorm:"column:updated_at;not null;default:now()"`
	RoleGroups   []RoleGroup `gorm:"many2many:admin_user_role_groups;joinForeignKey:AdminUserID;joinReferences:RoleGroupID"`
}

func (AdminUser) TableName() string {
	return "admin_users"
}

type RoleGroup struct {
	ID          int64                 `gorm:"column:id;primaryKey;autoIncrement"`
	Code        string                `gorm:"column:code;type:varchar(50);not null;uniqueIndex"`
	Name        string                `gorm:"column:name;type:varchar(100);not null"`
	Description *string               `gorm:"column:description;type:text"`
	IsActive    bool                  `gorm:"column:is_active;not null;default:true;index"`
	CreatedAt   time.Time             `gorm:"column:created_at;not null;default:now()"`
	UpdatedAt   time.Time             `gorm:"column:updated_at;not null;default:now()"`
	AdminUsers  []AdminUser           `gorm:"many2many:admin_user_role_groups;joinForeignKey:RoleGroupID;joinReferences:AdminUserID"`
	Permissions []RoleGroupPermission `gorm:"foreignKey:RoleGroupID;references:ID;constraint:OnDelete:CASCADE"`
}

func (RoleGroup) TableName() string {
	return "role_groups"
}

type AdminUserRoleGroup struct {
	AdminUserID int64     `gorm:"column:admin_user_id;primaryKey;not null"`
	RoleGroupID int64     `gorm:"column:role_group_id;primaryKey;not null"`
	CreatedAt   time.Time `gorm:"column:created_at;not null;default:now()"`
	AdminUser   AdminUser `gorm:"foreignKey:AdminUserID;references:ID;constraint:OnDelete:CASCADE"`
	RoleGroup   RoleGroup `gorm:"foreignKey:RoleGroupID;references:ID;constraint:OnDelete:CASCADE"`
}

func (AdminUserRoleGroup) TableName() string {
	return "admin_user_role_groups"
}

type RoleGroupPermission struct {
	RoleGroupID     int64     `gorm:"column:role_group_id;primaryKey;not null"`
	PermissionKey   string    `gorm:"column:permission_key;type:varchar(100);primaryKey;not null;index"`
	PermissionValue int64     `gorm:"column:permission_value;not null;default:0"`
	CreatedAt       time.Time `gorm:"column:created_at;not null;default:now()"`
	UpdatedAt       time.Time `gorm:"column:updated_at;not null;default:now()"`
	RoleGroup       RoleGroup `gorm:"foreignKey:RoleGroupID;references:ID;constraint:OnDelete:CASCADE"`
}

func (RoleGroupPermission) TableName() string {
	return "role_group_permissions"
}

type Menu struct {
	ID            int64     `gorm:"column:id;primaryKey;autoIncrement"`
	ParentID      int64     `gorm:"column:parent_id;not null;default:0;index"`
	Title         string    `gorm:"column:title;type:varchar(100);not null"`
	KeyCode       string    `gorm:"column:key_code;type:varchar(100);not null;uniqueIndex"`
	Icon          *string   `gorm:"column:icon;type:varchar(100)"`
	URL           *string   `gorm:"column:url;type:varchar(255)"`
	PermissionBit *int32    `gorm:"column:permission_bit;index"`
	IsActive      bool      `gorm:"column:is_active;not null;default:true;index"`
	SortOrder     int32     `gorm:"column:sort_order;not null;default:0;index"`
	CreatedAt     time.Time `gorm:"column:created_at;not null;default:now()"`
	UpdatedAt     time.Time `gorm:"column:updated_at;not null;default:now()"`
}

func (Menu) TableName() string {
	return "menus"
}

func AutoMigrateModels() []interface{} {
	return []interface{}{
		&AdminUser{},
		&RoleGroup{},
		&AdminUserRoleGroup{},
		&RoleGroupPermission{},
		&Menu{},
	}
}
