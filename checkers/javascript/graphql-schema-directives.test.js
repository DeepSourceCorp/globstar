// <expect-error>
const apollo_server_1 = new ApolloServer({
    typeDefs,
    resolvers,
    schemaDirectives: {
        rateLimit: rateLimitDirective
    },
});

// <no-error>
const apollo_server_3 = new ApolloServer({
    typeDefs,
    resolvers,
});