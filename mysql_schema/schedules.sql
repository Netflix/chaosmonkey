CREATE TABLE IF NOT EXISTS schedules (
    id INT NOT NULL AUTO_INCREMENT PRIMARY KEY,
    date         DATE NOT NULL,        -- date of termination schedule, in local time zone
    time         DATETIME NOT NULL,    -- time in UTC. Because of time difference, may differ from date
    app          VARCHAR(512) NOT NULL,
    account      VARCHAR(100) NOT NULL,
    region       VARCHAR(50)  NOT NULL, -- use blank string to indicate not present
    stack        VARCHAR(255) NOT NULL, -- use blank string to indicate not present
    cluster      VARCHAR(768) NOT NULL, -- use blank string to indicate not present
    INDEX date_index (date)
    )
ENGINE=InnoDB
