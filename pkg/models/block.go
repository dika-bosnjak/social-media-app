package models

import (
	"database/sql"
	"time"
)

type Block struct {
	ID            string    `json:"id" gorm:"primaryKey"`
	UserBlockID   string    `json:"user_block_id"  gorm:"type:varchar(191)"`
	UserBlockedID string    `json:"user_blocked_id"  gorm:"type:varchar(191)"`
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`
}

type BlockAPI struct {
	BlockID                 string `json:"block_id"`
	UserBlockID             string `json:"user_block_id"  gorm:"type:varchar(191)"`
	UserBlockedID           string `json:"user_blocked_id"  gorm:"type:varchar(191)"`
	UserBlockedFirstName    string `json:"user_blocked_firstname"`
	UserBlockedLastName     string `json:"user_blocked_lastname"`
	UserBlockedProfileImage string `json:"user_blocked_profile_image"`
}

func GetBlockedUsers(db *sql.DB, id string) ([]BlockAPI, error) {

	//get all blocked users
	var blocks []BlockAPI
	rows, err := db.Query(`SELECT blocks.id as BlockID, user_blocked_id, users.first_name, users.last_name, users.user_photo_url 
							FROM blocks 
							LEFT JOIN users ON users.id = blocks.user_blocked_id 
							WHERE user_block_id = ?`, id)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	//loop through the rows of the result and fullfill the blocks slice
	for rows.Next() {
		var block BlockAPI
		if err := rows.
			Scan(&block.BlockID,
				&block.UserBlockedID,
				&block.UserBlockedFirstName,
				&block.UserBlockedLastName,
				&block.UserBlockedProfileImage); err != nil {
			return blocks, err
		}
		blocks = append(blocks, block)
	}
	if err = rows.Err(); err != nil {
		return blocks, err
	}
	return blocks, nil
}

func GetBlockedUsersID(db *sql.DB, loggedInUserID string) ([]string, error) {

	//get all blocked users ids
	var blocked []string
	rows, err := db.Query(`SELECT user_blocked_id 
							FROM blocks 
							WHERE user_block_id = ? 
								UNION 
							SELECT user_block_id 
							FROM blocks 
							WHERE user_blocked_id = ? `, loggedInUserID, loggedInUserID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	//loop through the rows from the result and fullfill blocked slice
	for rows.Next() {
		var block string
		if err := rows.
			Scan(&block); err != nil {
			return blocked, err
		}
		blocked = append(blocked, block)
	}
	if err = rows.Err(); err != nil {
		return blocked, err
	}
	return blocked, nil
}

func CheckBlockStatus(db *sql.DB, loggedInUserID string, userID string) (bool, error) {

	//check whether the logged in user blocked or is blocked by the searched user
	rows, err := db.Query(`SELECT blocks.id 
								FROM blocks 
								LEFT JOIN users ON users.id = blocks.user_blocked_id 
								WHERE user_block_id = ? AND user_blocked_id = ?  
									UNION 
							SELECT blocks.id 
								FROM blocks 
								LEFT JOIN users ON users.id = blocks.user_blocked_id 
								WHERE user_block_id = ? AND user_blocked_id = ? `, loggedInUserID, userID, userID, loggedInUserID)
	if err != nil {
		return false, err
	}
	for rows.Next() {
		return true, nil
	}

	return false, nil
}

func NumberOfBlockedUsers(db *sql.DB, loggedInUserID string) int {

	//get the number of the blocked users
	blockedUsers, _ := GetBlockedUsers(db, loggedInUserID)
	return (len(blockedUsers))

}

func BlockUser(db *sql.DB, id string, userID string, blockedUserID string) {

	//block the user
	db.Exec(`INSERT INTO blocks (id, user_block_id, user_blocked_id) 
				VALUES (?, ?, ?)`, id, userID, blockedUserID)
}

func GetBlock(db *sql.DB, loggedInUserID string, userID string) (Block, error) {

	//get the info about the block
	var block Block
	if err := db.QueryRow(`SELECT * 
							FROM blocks  
							WHERE (user_block_id = ? AND user_blocked_id = ?) OR (user_blocked_id = ? AND user_block_id = ?)`, loggedInUserID, userID, loggedInUserID, userID).
		Scan(
			&block.ID,
			&block.UserBlockID,
			&block.UserBlockedID,
			&block.CreatedAt,
			&block.UpdatedAt); err != nil {
		if err == sql.ErrNoRows {
			return block, nil
		}

		return block, err
	}
	return block, nil
}

func Unblock(db *sql.DB, id string) error {

	//delete the block (unblock the person)
	_, err := db.Exec(`DELETE 
						FROM blocks 
						WHERE id = ?`, id)
	return err
}
