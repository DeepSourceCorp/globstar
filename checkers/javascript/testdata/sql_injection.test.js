// VULNERABLE PATTERNS

// Direct string concatenation/interpolation with user input
const v1 = "SELECT * FROM users WHERE username = '" + username + "'";
const v2 = `SELECT * FROM users WHERE email = '${email}' AND phone = ${phone}`;

// <expect-error>
connection.query(v1);

// <expect-error>
pool.query(v2);

// Variable defined elsewhere
// <expect-error>
connection.query(v3);

// ORMs with raw queries
// <expect-error>
sequelize.query(`SELECT * FROM products WHERE category = '${category}'`, {
  type: sequelize.QueryTypes.SELECT,
});

// <expect-error>
knex.raw(`SELECT * FROM users WHERE id = ${userId}`);

// <expect-error>
knex.raw(`SELECT * FROM users WHERE id = ${userId}` + `AND email = ${email}`);

// <expect-error>
const users = await prisma.$queryRawUnsafe(
  `SELECT * FROM ${table} WHERE id = ${id}`,
);

// <expect-error>
const result = await prisma.$executeRawUnsafe(
  `DELETE FROM users WHERE email = '${email}'`,
);

// SAFE PATTERNS

connection.query(s1, "SELECT * FROM user WHERE name LIKE 'A%'");

// Parameterized queries
const s1 = "SELECT * FROM users WHERE username = ?";
connection.query(s1, [username]);

const s2 = "SELECT * FROM users WHERE email = $1 AND phone = $2";
pool.query(s2, [email, phone]);

// Prepared statements
const preparedStatement = connection.prepare(
  "SELECT * FROM users WHERE id = ?",
);
preparedStatement.execute([userId]);
