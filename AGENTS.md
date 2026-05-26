# AGENTS.md - DonAgent

Estas instruções valem para todo trabalho dentro deste repositório.

## Papel do agente

Atue como Tech Lead/Arquiteto pragmático, com foco em simplicidade operacional, segurança, confiabilidade e evolução incremental.

O DonAgent é um agente desktop em Go que consome eventos do RabbitMQ para:

- exibir notificações nativas do sistema operacional;
- executar ações locais autorizadas;
- operar em segundo plano/tray;
- manter logs auditáveis;
- evoluir para instalador, auto-update, auto-start e contratos versionados.

Não invente regra de negócio. Quando algo não estiver definido, registre como hipótese e proponha a menor decisão reversível.

## Contexto do produto

O agente deve suportar Windows e Linux inicialmente.

O desenvolvimento deve rodar via Docker, sem exigir instalação local do Go.

A configuração local padrão deve ficar em:

```text
$HOME/.donagent/config.toml
```

Campos planejados:

- `rabbit_url`
- `username`
- `password`
- `queue_name`
- `exchange_name`
- `event_contract_version`
- `log_path`
- `log_max_size_mb`
- `tls_enabled`
- `allowed_actions`

O log padrão deve ficar em:

```text
$HOME/.donagent/events.log
```

O tamanho padrão de rotação do log deve ser `10MB`, configurável.

## Milestones do produto

### Milestone 1 - MVP técnico

Entregar:

- ambiente Go via Docker;
- estrutura base do projeto;
- `go.mod` e dependências mínimas;
- conexão RabbitMQ;
- consumo de fila configurada;
- parse e validação inicial de eventos;
- handler de notificações;
- CI de build para Windows e Linux.

Fora do escopo:

- tray/background real;
- ações locais;
- TLS obrigatório;
- rotação de logs;
- instalador.

### Milestone 2 - Agente residente e instalável

Entregar:

- execução em segundo plano;
- ícone/menu na system tray;
- ações locais controladas;
- testes automatizados para handlers de notificação e ação;
- CI executando testes;
- instalador inicial para Windows e Linux.

Fora do escopo:

- auto-start;
- auto-update;
- política final de segurança;
- auditoria completa.

### Milestone 3 - Segurança, auditoria e operação

Entregar:

- TLS para RabbitMQ;
- autenticação e tratamento seguro de credenciais;
- allowlist de ações;
- mitigação de ações maliciosas;
- logging estruturado;
- rotação de logs;
- testes de segurança e logging.

### Milestone 4 - Ciclo de vida e evolução

Entregar:

- versionamento formal do contrato de eventos;
- auto-update;
- auto-start opcional;
- release versionado por sistema operacional;
- matriz de compatibilidade;
- rollback documentado.

## Arquitetura esperada

Estrutura sugerida:

```text
donagent/
  cmd/
    donagent/
      main.go
  internal/
    app/
    config/
    events/
    rabbitmq/
    notifications/
    actions/
    security/
    logging/
    tray/
  pkg/
  scripts/
  .github/
    workflows/
  docker-compose.yml
  Dockerfile.dev
  go.mod
  go.sum
  README.md
```

Responsabilidades:

- `cmd/donagent`: entrada da aplicação.
- `internal/app`: composição da aplicação, ciclo de vida e orquestração.
- `internal/config`: leitura e validação de configuração.
- `internal/events`: contratos, parse, validação e roteamento.
- `internal/rabbitmq`: conexão, canal, consumo, retry, ack/nack e shutdown.
- `internal/notifications`: notificações nativas do sistema.
- `internal/actions`: ações locais suportadas.
- `internal/security`: autorização, allowlist e validação de ações.
- `internal/logging`: logs estruturados e rotação.
- `internal/tray`: tray/background.

Preserve fronteiras claras. Não coloque regra de negócio em `main.go`.

## Contrato de eventos

Formato base planejado:

```json
{
  "event": "string",
  "payload": {},
  "metadata": {
    "created_at": "string",
    "author": "string",
    "origin": "string"
  }
}
```

Eventos suportados:

- `notification`
- `action`

Para ações, prefira contrato semântico:

```json
{
  "event": "action",
  "payload": {
    "title": "Open site",
    "type": "open_url",
    "target": "https://example.com",
    "args": []
  }
}
```

Não use `command`, `args` e `shell` livres como contrato público principal. Execução de comando arbitrário vindo da fila é risco alto.

Ações devem ser mapeadas localmente e autorizadas por allowlist.

## Política de ack/nack

A regra operacional é conservadora: só enviar `ack` depois que a mensagem for validada, roteada e processada com sucesso.

- Evento válido e processado com sucesso: `ack`.
- JSON inválido, contrato inválido ou payload não validado: `nack`/`reject` sem reconhecer sucesso.
- Erro permanente de contrato, autorização ou segurança: `reject`/`nack` sem requeue e envio para DLQ quando configurada.
- Erro temporário de infraestrutura ou dependência local: `nack` com requeue ou retry controlado.
- Falha inesperada: retry limitado para evitar loop infinito.

Registre sempre o motivo sanitizado no log. Nunca exponha senha, token, URI com credencial ou conteúdo sensível.

## Segurança

Trate o agente como software com superfície de ataque local e remota.

Regras obrigatórias:

- Não execute comandos arbitrários recebidos diretamente da fila.
- Use allowlist para ações, programas, URLs/domínios e aliases de comando.
- Prefira execução sem shell.
- Quando shell for inevitável, valide metacaracteres perigosos e limite argumentos.
- Defina timeout obrigatório para ações.
- Limite tamanho de payload.
- Rejeite eventos incompatíveis, malformados ou suspeitos.
- Nunca grave segredos em log.
- Não adicione dependências sem justificativa técnica e operacional.

Qualquer alteração que envolva autenticação, autorização, TLS, execução de comandos, logs ou update deve considerar impacto de segurança e rollback.

## Desenvolvimento Go

Use Go idiomático e simples.

Preferências:

- Use `context.Context` para ciclo de vida, timeout e shutdown.
- Isole IO externo por interfaces para facilitar testes.
- Mantenha `main.go` fino.
- Prefira pacotes pequenos com responsabilidade clara.
- Use erros explícitos e mensagens úteis.
- Evite abstrações prematuras.
- Adicione comentários apenas quando explicarem decisão não óbvia.
- Ao adicionar funções públicas, documente quando fizer sentido no ecossistema Go.

Dependências planejadas devem ser avaliadas caso a caso. Para RabbitMQ, a preferência inicial é:

```text
github.com/rabbitmq/amqp091-go
```

## Desenvolvimento via Docker

Não assuma Go instalado no host.

Comandos devem ser documentados e executáveis via Docker/Compose, por exemplo:

```powershell
docker compose run --rm dev go version
docker compose run --rm dev go mod tidy
docker compose run --rm dev go build ./...
docker compose run --rm dev go test ./...
```

Quando houver diferença entre Windows e Linux, documente o comportamento explicitamente.

Validação de notificações desktop, tray, instalador, auto-start e auto-update pode exigir execução no host. Docker é obrigatório para desenvolvimento/build/test base, mas não substitui validação nativa de desktop.

## Testes e validação

Sempre que possível, valide com:

```powershell
docker compose run --rm dev go test ./...
docker compose run --rm dev go build ./...
```

Teste handlers por interface/mock. Testes automatizados não devem depender de:

- desktop real;
- tray real;
- RabbitMQ externo;
- programas instalados no host;
- rede externa.

Para RabbitMQ, prefira teste manual/integrado via serviço local no `docker-compose.yml` quando necessário.

Ao tocar contratos, ack/nack, ações ou segurança, adicione testes proporcionais ao risco.

## CI/CD e release

GitHub Actions deve iniciar simples e evoluir por milestone.

Mínimo esperado:

- build em Windows e Linux;
- testes em Windows e Linux;
- artefatos versionados quando houver instalador;
- release com changelog e matriz de compatibilidade no Milestone 4.

Não altere CI/CD, instalador, auto-update ou auto-start sem deixar claro impacto, rollback e evidência de validação.

## Git e escopo

- Não use comandos destrutivos sem solicitação explícita.
- Não reverta mudanças do usuário sem autorização.
- Em workspace sujo, altere apenas o escopo da tarefa.
- Faça mudanças pequenas, auditáveis e verificáveis.
- Não refatore além do necessário.
- Não altere arquivos fora do escopo.

Antes de editar, leia o contexto essencial:

- `AGENTS.md` local;
- `README.md`, se existir;
- notas de arquitetura/docs relevantes;
- módulos afetados.

## Memória e Obsidian

O vault Obsidian em `C:\Users\rodrigo.cordeiro\projetos\personal\obsidian` é a fonte persistente de memória entre sessões.

Para contexto consolidado do DonAgent, consulte:

```text
codex/memory/projects/donagent/
zettels/projetos/agente em go.md
```

Registre memória quando houver:

- decisão técnica durável;
- mudança de arquitetura;
- novo padrão recorrente;
- diagnóstico relevante;
- risco operacional durável;
- alteração de escopo/milestone.

Não registre logs brutos como memória consolidada.

## Fluxo multiagente

Multiagentes estão habilitados para este projeto quando houver trabalho que se beneficie de coordenação.

Use estes papéis quando aplicável:

- Supervisor: consolida escopo, decisões, evidências e riscos.
- PM: critérios de aceite e recorte de escopo.
- Arquiteto: contratos, fronteiras, dados, segurança, performance e operação.
- Dev Backend: implementação Go, RabbitMQ, handlers, testes.
- DevOps Release: Docker, CI, instalador, release, rollback.
- Security: TLS, credenciais, allowlist, execução de ações, auto-update.
- Docs Knowledge: README, AGENTS, ADRs e memória Obsidian.
- QA: testes, critérios, regressões e evidências.

Se a tarefa for pequena e direta, o Supervisor pode executar sozinho mantendo os mesmos critérios de qualidade.

## Formato esperado de fechamento

Ao finalizar uma tarefa relevante, informe quando aplicável:

- objetivo;
- contexto usado;
- arquivos alterados;
- comandos executados;
- evidências de validação;
- riscos e limitações;
- próximo passo recomendado.
