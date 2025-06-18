CREATE TABLE IF NOT EXISTS projects (
    id SERIAL PRIMARY KEY,
    name TEXT NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT now()
);

INSERT INTO projects (name, created_at)
SELECT 'Первая запись', now()
WHERE NOT EXISTS (SELECT 1 FROM projects);

CREATE TABLE IF NOT EXISTS goods (
    id SERIAL PRIMARY KEY,
    project_id INT NOT NULL,
    name TEXT NOT NULL,
    description TEXT,
    priority INT NOT NULL,
    removed BOOLEAN NOT NULL DEFAULT FALSE,
    created_at TIMESTAMP NOT NULL DEFAULT now(),

    FOREIGN KEY (project_id) REFERENCES projects(id) ON DELETE CASCADE
);

CREATE INDEX IF NOT EXISTS idx_goods_project_id ON goods(project_id);
CREATE INDEX IF NOT EXISTS idx_goods_priority ON goods(priority);
CREATE INDEX IF NOT EXISTS idx_goods_name ON goods(name);
CREATE INDEX IF NOT EXISTS idx_goods_removed ON goods(removed);
