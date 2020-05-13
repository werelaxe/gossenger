package dbapi

const createGetMaxTimeFunctionSql = `CREATE OR REPLACE FUNCTION GET_MAX_TIME(message_time TIMESTAMPTZ, chat_time TIMESTAMPTZ) RETURNS TIMESTAMP AS $$
	BEGIN
		RETURN (SELECT CASE WHEN message_time IS NULL THEN chat_time
			ELSE GREATEST(message_time, chat_time)
			END);
	END; $$
	LANGUAGE PLPGSQL;`

const createGetLastActionTimeFunctionSql = `CREATE OR REPLACE FUNCTION GET_LAST_ACTION_TIME(chat_refer_id integer) RETURNS TIMESTAMP AS $$
	BEGIN
	RETURN (SELECT GET_MAX_TIME(messages.created_at, chats.created_at)
		FROM MESSAGES RIGHT JOIN chats ON chat_refer=chats.id
		WHERE chats.id=chat_refer_id
		ORDER BY GET_MAX_TIME(messages.created_at, chats.created_at) DESC LIMIT 1);
	END; $$
	LANGUAGE PLPGSQL;`

func (api *Api) InitFunctions() {
	if err := api.Db.Exec(createGetMaxTimeFunctionSql).Error; err != nil {
		panic(err)
	}
	if err := api.Db.Exec(createGetLastActionTimeFunctionSql).Error; err != nil {
		panic(err)
	}
}
