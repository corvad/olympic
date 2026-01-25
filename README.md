# cascade

Simple link shortener api written in Go. Very early stages, not even close to ready.

Project was named after the Cascade Mountain Range.

## Deployment

### Fly.io Deployment

This project includes a `fly.toml` configuration file for easy deployment to [fly.io](https://fly.io).

#### Prerequisites
- Install the [flyctl CLI](https://fly.io/docs/hands-on/install-flyctl/)
- Sign up for a fly.io account

#### Deploy Steps

1. Launch your app (first time only):
```bash
fly launch
```

2. Set required secrets:
```bash
fly secrets set JWT_SECRET=your-secure-random-string-here
```

3. Create a volume for database persistence (first time only):
```bash
fly volumes create cascade_data --size 1
```

4. Deploy:
```bash
fly deploy
```

#### Configuration

The app requires the following environment variables:
- `JWT_SECRET`: Secret key for JWT token signing (set via fly secrets)
- `DB_FILE`: Database file path (configured in fly.toml as `/data/cascade.db`)

The app uses a persistent volume mounted at `/data` to store the SQLite database.

