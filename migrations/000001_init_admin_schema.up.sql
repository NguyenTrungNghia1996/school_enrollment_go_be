CREATE OR REPLACE FUNCTION set_updated_at()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = now();
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TABLE admin_users (
    id BIGSERIAL PRIMARY KEY,
    username VARCHAR(50) NOT NULL UNIQUE,
    password_hash VARCHAR(255) NOT NULL,
    full_name VARCHAR(100) NOT NULL,
    email VARCHAR(100) UNIQUE,
    phone_number VARCHAR(20),
    is_super_admin BOOLEAN NOT NULL DEFAULT FALSE,
    is_active BOOLEAN NOT NULL DEFAULT TRUE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE role_groups (
    id BIGSERIAL PRIMARY KEY,
    code VARCHAR(50) NOT NULL UNIQUE,
    name VARCHAR(100) NOT NULL,
    description TEXT,
    is_active BOOLEAN NOT NULL DEFAULT TRUE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE admin_user_role_groups (
    admin_user_id BIGINT NOT NULL,
    role_group_id BIGINT NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    PRIMARY KEY (admin_user_id, role_group_id),
    CONSTRAINT fk_admin_user_role_groups_admin_user
        FOREIGN KEY (admin_user_id) REFERENCES admin_users(id) ON DELETE CASCADE,
    CONSTRAINT fk_admin_user_role_groups_role_group
        FOREIGN KEY (role_group_id) REFERENCES role_groups(id) ON DELETE CASCADE
);

CREATE TABLE role_group_permissions (
    role_group_id BIGINT NOT NULL,
    permission_key VARCHAR(100) NOT NULL,
    permission_value BIGINT NOT NULL DEFAULT 0,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    PRIMARY KEY (role_group_id, permission_key),
    CONSTRAINT fk_role_group_permissions_role_group
        FOREIGN KEY (role_group_id) REFERENCES role_groups(id) ON DELETE CASCADE
);

CREATE TABLE menus (
    id BIGSERIAL PRIMARY KEY,
    parent_id BIGINT NOT NULL DEFAULT 0,
    title VARCHAR(100) NOT NULL,
    key_code VARCHAR(100) NOT NULL UNIQUE,
    icon VARCHAR(100),
    url VARCHAR(255),
    permission_bit INT,
    is_active BOOLEAN NOT NULL DEFAULT TRUE,
    sort_order INT NOT NULL DEFAULT 0,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX idx_admin_users_is_active
    ON admin_users (is_active);

CREATE INDEX idx_admin_users_is_super_admin
    ON admin_users (is_super_admin);

CREATE INDEX idx_role_groups_is_active
    ON role_groups (is_active);

CREATE INDEX idx_role_group_permissions_permission_key
    ON role_group_permissions (permission_key);

CREATE INDEX idx_menus_parent_id
    ON menus (parent_id);

CREATE INDEX idx_menus_permission_bit
    ON menus (permission_bit);

CREATE INDEX idx_menus_is_active
    ON menus (is_active);

CREATE INDEX idx_menus_sort_order
    ON menus (sort_order);

CREATE TRIGGER trg_admin_users_set_updated_at
    BEFORE UPDATE ON admin_users
    FOR EACH ROW
    EXECUTE FUNCTION set_updated_at();

CREATE TRIGGER trg_role_groups_set_updated_at
    BEFORE UPDATE ON role_groups
    FOR EACH ROW
    EXECUTE FUNCTION set_updated_at();

CREATE TRIGGER trg_role_group_permissions_set_updated_at
    BEFORE UPDATE ON role_group_permissions
    FOR EACH ROW
    EXECUTE FUNCTION set_updated_at();

CREATE TRIGGER trg_menus_set_updated_at
    BEFORE UPDATE ON menus
    FOR EACH ROW
    EXECUTE FUNCTION set_updated_at();
