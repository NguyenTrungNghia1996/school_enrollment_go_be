DROP TRIGGER IF EXISTS trg_menus_set_updated_at ON menus;
DROP TRIGGER IF EXISTS trg_role_group_permissions_set_updated_at ON role_group_permissions;
DROP TRIGGER IF EXISTS trg_role_groups_set_updated_at ON role_groups;
DROP TRIGGER IF EXISTS trg_admin_users_set_updated_at ON admin_users;

DROP TABLE IF EXISTS admin_user_role_groups;
DROP TABLE IF EXISTS role_group_permissions;
DROP TABLE IF EXISTS menus;
DROP TABLE IF EXISTS role_groups;
DROP TABLE IF EXISTS admin_users;

DROP FUNCTION IF EXISTS set_updated_at();
