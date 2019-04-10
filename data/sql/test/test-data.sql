INSERT INTO entities (pub_id) VALUES ('E9EB036A-0194-4AD4-B598-2412FB9C8F5B');
-- TODO: 'SET' is not ANSI SQL; for this and other reasons, we want to do a
-- replacement scheme. Possibly something like:
-- 1) Name template files with a commen prefix ('.sql.template').
-- 2) Use bash subsitutios, so "VALUES ($JANE_DOE_ID)"
-- 3) Have a 'template.vars' file.
-- 4) source template.vars; for $TEMPLATE in ...; do ...; eval "$(cat "$TEMPLATE")" > $SQL_FILE; done
SET @some_org_id=LAST_INSERT_ID();
INSERT INTO users (id, auth_id, active) VALUES (@some_org_id,'abcdefg123',0);
INSERT INTO orgs (id, display_name, summary, phone, email) VALUES (@some_org_id,'Some Org','Builders of things.','5555551111','janedoe@test.com');
