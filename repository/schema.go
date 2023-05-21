package repository

var schemas = [...]string{
	`CREATE TABLE IF NOT EXISTS users(
    	"id" TEXT PRIMARY KEY,
    	"name" TEXT NOT NULL,
    	"email" TEXT NOT NULL UNIQUE,
    	"enclave_url" TEXT NOT NULL,
    	"verification_key" TEXT NOT NULL);`,
	`CREATE TABLE IF NOT EXISTS wallets(
    	"address" TEXT PRIMARY KEY,
    	"name" TEXT NOT NULL,
    	"user_id" TEXT NOT NULL,
    	CONSTRAINT fk_user
            FOREIGN KEY("user_id")
    			REFERENCES users("id")
    			ON DELETE CASCADE,
    	UNIQUE("name", "user_id"));`,
	`CREATE TABLE IF NOT EXISTS contacts(
  		"user_id" TEXT NOT NULL,
  		"contact_id" TEXT NOT NULL,
  		PRIMARY KEY ("user_id", "contact_id"),
		CONSTRAINT fk_user
			FOREIGN KEY("user_id")
				REFERENCES users("id")
				ON DELETE CASCADE,
    	CONSTRAINT fk_contact
			FOREIGN KEY("contact_id")
				REFERENCES users("id")
				ON DELETE CASCADE);`,
	`CREATE TABLE IF NOT EXISTS signups(
  		"user_id" TEXT PRIMARY KEY,
  		"activated" BOOLEAN NOT NULL,
  		"activation_string" TEXT NOT NULL,
  		"created" BIGINT NOT NULL,
  		"valid_until" BIGINT NOT NULL,
		CONSTRAINT fk_user
			FOREIGN KEY("user_id")
				REFERENCES users("id")
				ON DELETE CASCADE);`,
	`CREATE TABLE IF NOT EXISTS notifications(
    	"id" BIGSERIAL,
    	"series_id" TEXT NOT NULL,
  		"creation_time" BIGINT NOT NULL,
  		"send_after" BIGINT NOT NULL,
  		"times_tried" BIGINT NOT NULL,
  		"user_id" TEXT NOT NULL,
  		"title" TEXT NOT NULL,
  		"body" TEXT NOT NULL,
		CONSTRAINT fk_user
			FOREIGN KEY("user_id")
				REFERENCES users("id")
				ON DELETE CASCADE);`,
}
