import sql from "k6/x/sql";
import oracle from "k6/x/sql/driver/oracle";

const connectionString = "oracle://app:app@localhost:1521/FREEPDB1";

export const options = {
  vus: 1,
  iterations: 1,
};

export default function () {
  const db = sql.open(oracle, connectionString, {
    max_open_conns: 1,
    max_idle_conns: 1,
  });

  db.exec(`
    BEGIN
      EXECUTE IMMEDIATE 'DROP TABLE trace_test PURGE';
    EXCEPTION
      WHEN OTHERS THEN
        IF SQLCODE != -942 THEN
          RAISE;
        END IF;
    END;
  `);

  db.exec(`
    CREATE TABLE trace_test (
      id NUMBER GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
      name VARCHAR2(100) NOT NULL
    )
  `);

  db.exec("INSERT INTO trace_test (name) VALUES (:1)", "hello");

  const rows = db.query("SELECT id, name FROM trace_test ORDER BY id");
  console.log(JSON.stringify(rows));

  db.exec("DROP TABLE trace_test PURGE");
  db.close();
}
