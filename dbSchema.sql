CREATE DATABASE IF NOT EXISTS karen;
USE karen;

CREATE TABLE IF NOT EXISTS users (
    ID INTEGER NOT NULL AUTO_INCREMENT,
    Email VARCHAR(255) UNIQUE,
    Name VARCHAR(255),
    Password VARCHAR(255),
    AvatarURL VARCHAR(255),
    PRIMARY KEY(ID)
);
