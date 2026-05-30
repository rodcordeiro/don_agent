# Backlog - DonAgent

## Objetivo

Organizar o backlog do DonAgent em sprints pequenas, verificáveis e alinhadas aos milestones do produto:

- M1: MVP técnico com Go via Docker, RabbitMQ, eventos, notificações e CI de build.
- M2: agente residente, tray, identidade visual, ações locais, testes e instalador.
- M3: segurança, auditoria, logging e hardening operacional.
- M4: versionamento de contrato, auto-update, auto-start e release controlado.

## Premissas

- Desenvolvimento base deve rodar via Docker, sem exigir Go instalado no host.
- Windows e Linux são plataformas-alvo iniciais.
- O agente consome eventos do RabbitMQ.
- O contrato público não deve permitir comando arbitrário livre vindo da fila.
- `ack` só deve ocorrer após validação, roteamento e processamento bem-sucedidos.
- Notificações desktop podem exigir validação nativa fora do Docker.
- Cada sprint tem no máximo 5 tarefas.

## Prioridades

| Prioridade | Significado |
|---|---|
| P0 | Necessário para desbloquear o milestone atual |
| P1 | Importante para completar o milestone com qualidade |
| P2 | Pode entrar após o fluxo principal estar funcional |

## Plano de execução do Milestone 1

### Objetivo do M1

Entregar um agente Go mínimo, executável via Docker, capaz de:

- carregar configuração local;
- conectar ao RabbitMQ;
- consumir mensagens de uma fila configurada;
- validar contrato inicial de eventos;
- processar evento `notification`;
- aplicar política básica de `ack`/`nack`;
- compilar em CI para Windows e Linux.

### Sequência recomendada do M1

1. Criar base Docker + projeto Go.
2. Criar configuração mínima e estrutura interna.
3. Implementar contrato de eventos e validação.
4. Implementar RabbitMQ consumer com `ack`/`nack`.
5. Implementar handler de notificação por interface.
6. Fechar CI de build.
7. Documentar execução local e teste manual com RabbitMQ.

### Critérios de conclusão do M1

- `docker compose run --rm dev go build ./...` executa com sucesso.
- Agente inicia lendo configuração mínima.
- RabbitMQ local sobe via Docker Compose.
- Agente consome mensagem válida da fila configurada.
- Evento `notification` válido é roteado para handler de notificação.
- Evento inválido não derruba o agente.
- Mensagem válida processada com sucesso recebe `ack`.
- Mensagem inválida ou contrato rejeitado não recebe `ack` de sucesso.
- GitHub Actions executa build para Windows e Linux.
- Fora do escopo: tray, ações locais, TLS obrigatório, rotação de logs e instalador.

## Sprints

### Sprint 1 - Bootstrap do projeto e ambiente Docker

Milestone: M1

Prioridade: P0

Objetivo: criar a base mínima para desenvolvimento Go via Docker.

#### Tarefas

1. Criar `go.mod` e estrutura base do projeto.
2. Criar `Dockerfile.dev` com imagem oficial Go.
3. Criar `docker-compose.yml` com serviço `dev`.
4. Criar `cmd/donagent/main.go` fino, delegando composição para `internal/app`.
5. Criar README inicial com comandos Docker.

#### Critérios de aceite

- `docker compose run --rm dev go version` funciona.
- `docker compose run --rm dev go build ./...` funciona.
- `main.go` não contém regra de negócio.
- Estrutura inicial respeita fronteiras `cmd/` e `internal/`.

### Sprint 2 - Configuração local

Milestone: M1

Prioridade: P0

Objetivo: permitir inicialização previsível por arquivo de configuração.

#### Tarefas

1. Criar pacote `internal/config`.
2. Definir leitura padrão de `$HOME/.donagent/config.toml`.
3. Criar `config.example.toml`.
4. Validar campos mínimos: `rabbit_url`, `queue_name`, `exchange_name`.
5. Aplicar defaults para `log_path`, `log_max_size_mb` e `event_contract_version`.

#### Critérios de aceite

- Agente falha com erro claro quando configuração obrigatória está ausente.
- Agente inicia com configuração válida.
- Defaults documentados batem com `AGENTS.md`.
- Nenhum segredo é impresso em log ou console.

### Sprint 3 - Contrato inicial de eventos

Milestone: M1

Prioridade: P0

Objetivo: criar parse e validação mínima para eventos recebidos.

#### Tarefas

1. Criar pacote `internal/events`.
2. Criar modelo base `Event`, `Metadata` e `NotificationPayload`.
3. Definir constantes para `notification` e `action`.
4. Implementar parse JSON com validação de tipo de evento.
5. Implementar erros explícitos para JSON inválido, evento desconhecido e payload inválido.

#### Critérios de aceite

- Evento `notification` válido é parseado.
- JSON inválido retorna erro controlado.
- Evento desconhecido retorna erro controlado.
- Payload inválido não segue para processamento.
- Testes unitários cobrem parse válido e inválido.

### Sprint 4 - RabbitMQ local e consumer

Milestone: M1

Prioridade: P0

Objetivo: conectar e consumir mensagens do RabbitMQ.

#### Tarefas

1. Adicionar serviço RabbitMQ no `docker-compose.yml`.
2. Adicionar dependência `github.com/rabbitmq/amqp091-go`.
3. Criar pacote `internal/rabbitmq`.
4. Implementar conexão, canal e consumo da fila configurada.
5. Implementar shutdown com `context.Context`.

#### Critérios de aceite

- RabbitMQ sobe via Docker Compose.
- Agente conecta usando configuração local.
- Agente consome mensagens da fila configurada.
- Falha de conexão gera erro claro.
- Encerramento não deixa goroutines presas.

### Sprint 5 - Ack/Nack e roteamento de eventos

Milestone: M1

Prioridade: P0

Objetivo: garantir processamento conservador e auditável.

#### Tarefas

1. Criar roteador de eventos em `internal/events` ou `internal/app`.
2. Encaminhar `notification` para handler dedicado.
3. Aplicar `ack` somente após processamento bem-sucedido.
4. Aplicar `nack/reject` sem requeue para contrato inválido.
5. Aplicar `nack` com requeue apenas para erro temporário classificado.

#### Critérios de aceite

- Evento válido processado com sucesso recebe `ack`.
- JSON inválido não recebe `ack`.
- Evento desconhecido é rejeitado sem derrubar o agente.
- Erros permanentes e temporários têm tratamento distinto.
- Motivo sanitizado é registrado.

### Sprint 6 - Handler de notificação

Milestone: M1

Prioridade: P0

Objetivo: processar evento `notification` sem acoplar domínio à biblioteca nativa.

#### Tarefas

1. Criar pacote `internal/notifications`.
2. Definir interface `Notifier`.
3. Implementar handler `notification`.
4. Escolher biblioteca cross-platform de notificação desktop.
5. Tratar falha de notificação como erro de processamento controlado.

#### Critérios de aceite

- Handler recebe `title`, `description` e `image_url`.
- Campos opcionais são tratados sem crash.
- Implementação real fica atrás da interface.
- Teste automatizado usa mock de `Notifier`.
- Decisão da biblioteca registra trade-offs, limitações por sistema operacional e justificativa da dependência.
- Validação nativa é documentada como manual quando necessário.

### Sprint 7 - CI de build do M1

Milestone: M1

Prioridade: P0

Objetivo: garantir build mínimo em Windows e Linux.

#### Tarefas

1. Criar workflow GitHub Actions.
2. Configurar matriz `windows-latest` e `ubuntu-latest`.
3. Executar `go mod download`.
4. Executar `go build ./...`.
5. Documentar status esperado do CI no README.

#### Critérios de aceite

- Push/PR dispara workflow.
- Build passa em Windows e Linux.
- Falha de build bloqueia o workflow.
- Workflow não depende de RabbitMQ externo.

### Sprint 8 - Documentação e fechamento do M1

Milestone: M1

Prioridade: P1

Objetivo: tornar o MVP reproduzível por outro desenvolvedor.

#### Tarefas

1. Documentar setup local via Docker.
2. Documentar configuração em `$HOME/.donagent/config.toml`.
3. Documentar como subir RabbitMQ local.
4. Documentar payload de teste `notification`.
5. Documentar limitações do M1.

#### Critérios de aceite

- Um desenvolvedor consegue executar build seguindo o README.
- Um desenvolvedor consegue subir RabbitMQ local.
- Um payload de exemplo pode ser publicado e consumido.
- Limitações de tray, ações, TLS, logs e instalador estão explícitas.

### Sprint 9 - Tray e execução residente

Milestone: M2

Prioridade: P1

Objetivo: transformar o agente em aplicação residente com identificação visual consistente.

#### Tarefas

1. Escolher biblioteca de tray cross-platform.
2. Criar pacote `internal/tray`.
3. Exibir status básico de conexão.
4. Adicionar ações de pausar, retomar e sair.
5. Aplicar `assets/logo.png` como ícone do agente na tray/agente instalado.
6. Garantir encerramento gracioso pela tray.

#### Critérios de aceite

- Agente permanece em segundo plano.
- Usuário consegue sair pela tray.
- Pausar/retomar afeta consumo sem matar processo.
- Tray/agente instalado usam a identidade visual do DonAgent.
- Diferenças Windows/Linux ficam documentadas.

### Sprint 10 - Contrato e execução de ações locais

Milestone: M2

Prioridade: P1

Objetivo: suportar ações semânticas controladas.

#### Tarefas

1. Criar `ActionPayload`.
2. Criar interface `ActionExecutor`.
3. Implementar roteamento por `type`.
4. Suportar `open_url`.
5. Suportar `open_app` por alias/configuração.

#### Critérios de aceite

- Evento `action` válido é roteado corretamente.
- Tipo desconhecido é rejeitado.
- `open_url` abre URL válida.
- `open_app` só executa alvo configurado.
- Nenhum comando arbitrário livre é aceito.

### Sprint 11 - Testes automatizados e CI de testes

Milestone: M2

Prioridade: P1

Objetivo: cobrir handlers e contratos sem depender do desktop real.

#### Tarefas

1. Testar parse e validação de eventos.
2. Testar handler de notificação com mock.
3. Testar handler de ação com mock.
4. Testar rejeição de payload inválido.
5. Atualizar CI para `go test ./...`.

#### Critérios de aceite

- Testes passam via Docker.
- Testes passam no CI em Windows e Linux.
- Testes não dependem de RabbitMQ externo.
- Testes não dependem de tray ou desktop real.

### Sprint 12 - Release portátil via GitHub Actions

Milestone: M2

Prioridade: P1

Objetivo: gerar artefatos portáteis baixáveis no GitHub Actions antes do instalador nativo.

#### Tarefas

1. Criar workflow GitHub Actions para release portátil.
2. Reaproveitar `scripts/package-portable.ps1` ou fluxo equivalente no CI.
3. Gerar pacote Windows `donagent-<version>-windows-amd64.zip`.
4. Gerar pacote Linux `donagent-<version>-linux-amd64.tar.gz`.
5. Publicar artefatos do workflow contendo binário, `config.example.toml` e `assets/logo.png`.

#### Critérios de aceite

- Workflow pode ser disparado manualmente por `workflow_dispatch`.
- Artefatos Windows e Linux ficam disponíveis para download no GitHub Actions.
- Workflow executa testes antes de empacotar.
- Artefatos incluem `config.example.toml` e `assets/logo.png`.
- Workflow não promete instalação, auto-start, assinatura, auto-update ou uninstall.
- Falha de build/test impede publicação dos artefatos.

### Sprint 13 - Instalador inicial

Milestone: M2

Prioridade: P1

Objetivo: gerar distribuição inicial instalável ou empacotável com identidade visual do DonAgent.

#### Tarefas

1. Definir estratégia de pacote Windows.
2. Definir estratégia de pacote Linux.
3. Gerar binários versionados.
4. Incluir `config.example.toml`.
5. Incluir `assets/logo.png` e gerar formatos/tamanhos necessários para instalador, binário instalado e notificações nativas.
6. Documentar instalação e desinstalação.

#### Critérios de aceite

- Artefato Windows é gerado.
- Artefato Linux é gerado.
- Instalação preserva configuração local existente.
- Desinstalação não remove logs/config sem ação explícita.
- Instalador, agente instalado/tray e notificações do sistema operacional usam o logo do DonAgent para identificação consistente.

### Sprint 14 - TLS e credenciais

Milestone: M3

Prioridade: P0

Objetivo: endurecer comunicação e tratamento de segredos.

#### Tarefas

1. Adicionar configuração de TLS.
2. Suportar CA certificate.
3. Suportar credenciais por variável de ambiente.
4. Mascarar segredos em logs.
5. Documentar configuração segura.

#### Critérios de aceite

- TLS funciona quando habilitado.
- Certificado inválido bloqueia conexão.
- Segredos não aparecem em logs.
- Fallback inseguro não ocorre silenciosamente.

### Sprint 15 - Higiene de exposição e dados sensíveis

Milestone: M3

Prioridade: P0

Objetivo: corrigir riscos de exposição acidental identificados em revisão Sentinel antes de ampliar uso operacional.

#### Regra operacional específica

- Antes da execução, Sentinel deve revisar novamente o projeto para confirmar riscos, escopo e prioridades de segurança.
- PM deve auxiliar na estruturação do backlog interno da sprint, quebrando achados em tarefas executáveis e critérios de aceite.
- Esta sprint é exceção à premissa geral de limite máximo de 5 tarefas, porque pode consolidar múltiplos achados de segurança relacionados.

#### Tarefas

1. Restringir portas do RabbitMQ local no `docker-compose.yml` para bind em `127.0.0.1`.
2. Adicionar ignores preventivos para logs e configurações locais, como `*.log`, `events.log`, `config.toml` e `*.local.toml`.
3. Evitar saída bruta de `title`, `description` e `image_url` em console/logs operacionais.
4. Sanitizar ou truncar URLs e campos potencialmente sensíveis antes de registrar falhas.
5. Documentar que `guest/guest` existe apenas para ambiente local de desenvolvimento.

#### Critérios de aceite

- RabbitMQ local não fica exposto fora de loopback por padrão.
- Logs e configs locais comuns não aparecem como arquivos versionáveis por acidente.
- Conteúdo bruto de notificação não é impresso em logs operacionais.
- URLs com credenciais, tokens ou parâmetros sensíveis são mascaradas ou truncadas.
- Exemplos `guest/guest` continuam claramente marcados como credenciais locais, não produção.

### Sprint 16 - Allowlist e segurança de ações

Milestone: M3

Prioridade: P0

Objetivo: impedir execução maliciosa por eventos remotos.

#### Tarefas

1. Criar allowlist de ações.
2. Criar allowlist de programas.
3. Criar allowlist de domínios/URLs.
4. Restringir `run_command` a aliases.
5. Bloquear metacaracteres perigosos quando aplicável.

#### Critérios de aceite

- Ação fora da allowlist é recusada.
- Comando arbitrário vindo da fila não é executado.
- Payload suspeito é rejeitado.
- Rejeições são registradas com motivo sanitizado.

### Sprint 17 - Logging estruturado e rotação

Milestone: M3

Prioridade: P1

Objetivo: criar trilha auditável local.

#### Tarefas

1. Criar pacote `internal/logging`.
2. Gravar logs em `$HOME/.donagent/events.log`.
3. Registrar tipo, origem, autor, resultado e duração.
4. Implementar rotação por tamanho.
5. Aplicar default de `10MB`.

#### Critérios de aceite

- Logs são gravados em arquivo.
- Rotação ocorre ao atingir limite configurado.
- Erros são sanitizados.
- Teste cobre rotação com limite reduzido.

### Sprint 18 - Versionamento do contrato

Milestone: M4

Prioridade: P0

Objetivo: controlar evolução do contrato de eventos.

#### Tarefas

1. Formalizar campo de versão do contrato.
2. Documentar versões suportadas.
3. Validar versão antes do roteamento.
4. Definir política para versão desconhecida.
5. Criar testes para versões suportadas e rejeitadas.

#### Critérios de aceite

- Contrato incompatível é rejeitado.
- Versão suportada é processada.
- Política de compatibilidade está documentada.
- Produtores têm guia de migração.

### Sprint 19 - Auto-start e single instance

Milestone: M4

Prioridade: P2

Objetivo: permitir inicialização automática controlada.

#### Tarefas

1. Implementar auto-start no Windows.
2. Implementar auto-start no Linux.
3. Permitir habilitar/desabilitar configuração.
4. Implementar proteção contra múltiplas instâncias.
5. Documentar rollback manual.

#### Critérios de aceite

- Agente inicia automaticamente quando habilitado.
- Auto-start pode ser desabilitado.
- Segunda instância não consome fila em paralelo.
- Comportamento por sistema operacional está documentado.

### Sprint 20 - Auto-update e release operacional

Milestone: M4

Prioridade: P2

Objetivo: preparar ciclo de vida de atualização e rollback.

#### Tarefas

1. Definir origem confiável de updates.
2. Validar checksum ou assinatura de artefato.
3. Implementar atualização preservando config/logs.
4. Implementar rollback em falha.
5. Gerar release com changelog e matriz de compatibilidade.

#### Critérios de aceite

- Artefato inválido é recusado.
- Atualização preserva configuração local.
- Falha de update não inutiliza o agente.
- Release documenta compatibilidade e rollback.

## Riscos

- Notificação e tray podem ter diferenças relevantes entre Windows e Linux.
- Docker valida build/test base, mas não substitui validação nativa de desktop.
- `run_command` é superfície de alto risco e deve permanecer restrito.
- `ack/nack` incorreto pode causar perda de mensagens ou loop de reprocessamento.
- Auto-update só deve entrar após logging, segurança e instalador estarem estáveis.

## Próximo passo recomendado

Iniciar pela Sprint 1 e só avançar para RabbitMQ depois que o projeto compilar via Docker. O caminho crítico do M1 é: ambiente Go via Docker -> configuração -> contrato de eventos -> consumo RabbitMQ -> ack/nack -> notificação -> CI.
