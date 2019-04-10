CREATE TABLE `orgs` (
  `id` INT(10),
  `display_name` VARCHAR(128),
  `summary` VARCHAR(512),
-- see ../docs/Relational-Schemas.md#reformatting-data-via-a-trigger
  `phone` VARCHAR(12),
  `email` VARCHAR(255) NOT NULL,
  `homepage` VARCHAR(255),
  `logo_url` VARCHAR(255),
  CONSTRAINT `orgs_key` PRIMARY KEY ( `id` ),
  CONSTRAINT `orgs_ref_users` FOREIGN KEY ( `id` ) REFERENCES `users` ( `id` )
);
DELIMITER //
CREATE TRIGGER `persons_phone_format`
  BEFORE INSERT ON persons FOR EACH ROW
    BEGIN
      SET new.phone=(SELECT NUMERIC_ONLY(new.phone));
    END;//
DELIMITER ;
