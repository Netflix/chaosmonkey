CREATE TABLE IF NOT EXISTS terminations (
    id INT NOT NULL AUTO_INCREMENT PRIMARY KEY,
    app          VARCHAR(512) NOT NULL,
    account      VARCHAR(100) NOT NULL,
    stack        VARCHAR(255) NOT NULL,
    cluster      VARCHAR(768) NOT NULL,
    region       VARCHAR(50) NOT NULL,
    asg          VARCHAR(1000) NOT NULL,
    instance_id  VARCHAR(48) NOT NULL,
    killed_at    DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    leashed      BOOLEAN NOT NULL DEFAULT FALSE,
    INDEX app_killed_at_index (app,killed_at)
    )
ENGINE=InnoDB
