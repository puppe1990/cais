# Deploy no AWS Lightsail (systemd + Caddy)

Guia para instâncias Ubuntu nano (512MB RAM) sem Docker — o padrão usado em apps Cais com Caddy na mesma máquina.

## 1. Build na máquina de desenvolvimento

```bash
cais css
cais build --os linux --arch amd64 -o bin/server-linux
cais doctor   # valida web/static + manifest
```

Empacote binário + static:

```bash
tar czf release.tar.gz bin/server-linux web/static
scp -i ~/.ssh/your-key.pem release.tar.gz ubuntu@SEU_IP:/tmp/
```

## 2. Layout no servidor

```text
/opt/myapp/
  current/
    bin/server          # binário renomeado de server-linux
    web/static/         # CSS, JS, manifest, icons
  data/app.db           # SQLite persistente
/etc/myapp/env          # variáveis de produção (chmod 600)
```

`systemd` **WorkingDirectory** deve ser `/opt/myapp/current` (onde existe `web/static/`), ou defina `STATIC_DIR` no `.env`.

## 3. Variáveis de produção

```bash
PORT=:4006
ENV=production
APP_URL=https://myapp.example.com
DB_PATH=/opt/myapp/data/app.db
ADMIN_TOKEN=<token-forte>
LOCALE=pt
TRUSTED_PROXIES=127.0.0.1
# STATIC_DIR=/opt/myapp/current/web/static   # opcional
```

## 4. systemd

Copie `deploy/systemd/cais-app.service.example` e ajuste nomes/caminhos:

```bash
sudo systemctl daemon-reload
sudo systemctl enable myapp
sudo systemctl start myapp
curl -s http://127.0.0.1:4006/health
```

## 5. Caddy (TLS automático)

```
myapp.example.com {
  encode gzip
  reverse_proxy 127.0.0.1:4006
}
```

Aponte DNS `A` para o IP estático da instância. Caddy emite o certificado após propagação.

## 6. Dados iniciais

Seeds de desenvolvimento (usuário demo) **não** rodam em `ENV=production`. Para catálogo idempotente:

```bash
cais db seed
```

Registre seeds de domínio em `internal/db/seeds.go` — seguros para rodar em produção.

## 7. Verificação

```bash
curl -sI https://myapp.example.com/ | grep -i permissions-policy
curl -s https://myapp.example.com/static/manifest.webmanifest | grep display
curl -s https://myapp.example.com/health
```

## Troubleshooting

| Sintoma                        | Correção                                              |
| ------------------------------ | ----------------------------------------------------- |
| `web/static not found` no boot | `WorkingDirectory=/opt/myapp/current` ou `STATIC_DIR` |
| Certificado TLS pendente       | DNS `A` ainda não propagou                            |
| App vazio após deploy          | `cais db seed` ou `/signup`                           |
| IP errado nos logs             | `TRUSTED_PROXIES=127.0.0.1` atrás do Caddy            |
