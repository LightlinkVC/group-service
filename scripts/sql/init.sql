CREATE TABLE IF NOT EXISTS roles (
    id SERIAL PRIMARY KEY, 
    name VARCHAR(255) NOT NULL UNIQUE 
);

CREATE TABLE IF NOT EXISTS group_types (
    id SERIAL PRIMARY KEY, 
    name VARCHAR(255) NOT NULL UNIQUE
);

CREATE TABLE IF NOT EXISTS groups (
    id SERIAL PRIMARY KEY,
    name VARCHAR(255) NOT NULL,
    creator_id INTEGER NOT NULL,
    type_id INTEGER NOT NULL,
    CONSTRAINT fk_group_type FOREIGN KEY (type_id) REFERENCES group_types(id)
);

CREATE TABLE IF NOT EXISTS group_members (
    user_id INTEGER NOT NULL,
    group_id INTEGER NOT NULL,
    role_id INTEGER NOT NULL,
    CONSTRAINT pk_group_member PRIMARY KEY (user_id, group_id),
    CONSTRAINT fk_group FOREIGN KEY (group_id) REFERENCES groups(id),
    CONSTRAINT fk_role FOREIGN KEY (role_id) REFERENCES roles(id)
);

INSERT INTO roles (name) VALUES
    ('admin'),
    ('member')
ON CONFLICT (name) DO NOTHING;

INSERT INTO group_types (name) VALUES
    ('personal'),
    ('group')
ON CONFLICT (name) DO NOTHING;

CREATE TABLE IF NOT EXISTS messages (
    id SERIAL PRIMARY KEY,
    user_id INTEGER NOT NULL,
    group_id INTEGER NOT NULL,
    content TEXT NOT NULL,
    CONSTRAINT fk_message_user FOREIGN KEY (user_id, group_id) REFERENCES group_members(user_id, group_id),
    CONSTRAINT fk_message_group FOREIGN KEY (group_id) REFERENCES groups(id)
);
