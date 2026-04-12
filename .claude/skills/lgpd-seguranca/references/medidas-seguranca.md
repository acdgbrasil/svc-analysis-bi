# Medidas de Segurança - LGPD

## Sumário
1. [Fundamentação Legal (Art. 46)](#fundamentação-legal)
2. [Medidas Técnicas](#medidas-técnicas)
3. [Medidas Administrativas](#medidas-administrativas)
4. [Privacy by Design - 7 Princípios](#privacy-by-design)
5. [Privacy by Default](#privacy-by-default)
6. [Ciclo de Vida dos Dados e Ativos](#ciclo-de-vida)
7. [Controles Prioritários por Fase](#controles-por-fase)

---

## Fundamentação Legal (Art. 46)

**Art. 46 da LGPD**: "Os agentes de tratamento devem adotar medidas de segurança, técnicas
e administrativas aptas a proteger os dados pessoais de acessos não autorizados e de
situações acidentais ou ilícitas de destruição, perda, alteração, comunicação ou qualquer
forma de tratamento inadequado ou ilícito."

**Art. 46, § 1º**: As medidas devem ser observadas desde a fase de concepção do produto/serviço
até sua execução (**Privacy by Design**).

**Art. 46, § 2º**: As medidas serão regulamentadas pela ANPD, que poderá dispor sobre padrões
técnicos mínimos, considerando a natureza das informações, as características do tratamento
e o estado atual da tecnologia.

**Art. 47**: Os agentes de tratamento ou qualquer pessoa que intervenha em uma das fases do
tratamento obriga-se a garantir a segurança da informação, mesmo após o término do tratamento.

---

## Medidas Técnicas

### Controle de Acesso
- Autenticação multifator (MFA) para sistemas com dados pessoais
- Princípio do menor privilégio (acesso mínimo necessário)
- Controle de acesso baseado em papéis (RBAC)
- Revisão periódica de permissões (trimestral recomendado)
- Gestão de identidades e acessos (IAM)
- Bloqueio automático após tentativas falhas de login
- Segregação de funções (SoD)

### Criptografia
- **Em repouso**: AES-256 ou equivalente para dados armazenados
- **Em trânsito**: TLS 1.3+ para todas as comunicações
- **Gestão de chaves**: rotação periódica, armazenamento seguro (HSM recomendado)
- **Hashing**: senhas com bcrypt/scrypt/Argon2 + salt
- **Tokenização**: para dados sensíveis em ambientes de desenvolvimento/teste

### Proteção de Rede
- Firewalls com regras restritivas
- Segmentação de rede (dados pessoais em VLAN separada)
- IDS/IPS (Intrusion Detection/Prevention Systems)
- VPN para acessos remotos
- Monitoramento de tráfego (SIEM)
- WAF (Web Application Firewall) para aplicações web

### Proteção de Endpoints
- Antivírus/EDR atualizado em todos os dispositivos
- Criptografia de disco completo (BitLocker, FileVault)
- Políticas de atualização de patches (máximo 30 dias para patches críticos)
- DLP (Data Loss Prevention) para dados pessoais
- Gestão de dispositivos móveis (MDM)

### Backup e Recuperação
- Backup regular (diário para dados críticos)
- Criptografia dos backups
- Armazenamento em local seguro (offsite ou nuvem)
- Testes periódicos de restauração (mensal recomendado)
- Plano de Recuperação de Desastres (DRP)

### Monitoramento e Logs
- Logging de todos os acessos a dados pessoais
- Monitoramento em tempo real (SIEM)
- Retenção de logs por período adequado (mínimo 6 meses, recomendado 1 ano)
- Proteção contra adulteração de logs (write-once)
- Alertas automatizados para atividades suspeitas

### Desenvolvimento Seguro
- Ciclo de desenvolvimento seguro (SDLC)
- Revisão de código com foco em segurança
- Testes de penetração periódicos (anual no mínimo)
- Análise de vulnerabilidades automatizada (SAST/DAST)
- Sanitização de dados em ambientes de teste

---

## Medidas Administrativas

### Políticas e Procedimentos
- Política de Segurança da Informação
- Política de Privacidade e Proteção de Dados
- Política de Classificação da Informação
- Política de Uso Aceitável de Recursos de TI
- Procedimento de Resposta a Incidentes
- Política de Retenção e Descarte de Dados
- Política de Gestão de Terceiros

### Treinamento e Conscientização
- Treinamento obrigatório para todos os colaboradores (anual)
- Treinamento específico para áreas que tratam dados pessoais
- Campanhas de conscientização sobre phishing
- Simulações de incidentes de segurança
- Registro de participação em treinamentos

### Gestão de Terceiros
- Due diligence de segurança em fornecedores
- Cláusulas contratuais de proteção de dados (Art. 39)
- Auditorias periódicas de fornecedores
- Acordos de nível de serviço (SLA) de segurança
- Notificação obrigatória de incidentes por terceiros

### Gestão de Riscos
- Avaliação periódica de riscos de segurança (anual no mínimo)
- Metodologia documentada (ISO 27005, ISO 31000 ou NIST RMF)
- Registro de riscos com tratamento definido
- Revisão de riscos após incidentes ou mudanças significativas

---

## Privacy by Design - 7 Princípios

Conceito criado por Ann Cavoukian, incorporado ao Art. 46, § 1º da LGPD:

### 1. Proativo, não reativo; preventivo, não corretivo
- Antecipar e prevenir eventos de invasão de privacidade antes que ocorram
- Não esperar riscos se materializarem para então agir
- Monitoramento contínuo e avaliação proativa

### 2. Privacidade como padrão (Privacy by Default)
- Dados pessoais automaticamente protegidos em qualquer sistema
- Usuário não precisa agir para ter sua privacidade garantida
- Coletar apenas o mínimo necessário por padrão
- Compartilhamento desabilitado por padrão

### 3. Privacidade incorporada ao design
- Integrar proteção de dados na arquitetura de sistemas
- Não como complemento posterior (add-on)
- Considerada desde a fase de planejamento
- Componente integral da funcionalidade

### 4. Funcionalidade total (soma positiva)
- Evitar falsos dilemas (privacidade vs. funcionalidade)
- Buscar soluções que atendam todos os interesses legítimos
- Privacidade não deve prejudicar a experiência do usuário
- Abordagem "ganha-ganha"

### 5. Segurança ponta a ponta (ciclo de vida completo)
- Proteção durante todo o ciclo de vida dos dados
- Desde a coleta até a eliminação
- Criptografia, controle de acesso, integridade
- Destruição segura ao final do ciclo

### 6. Visibilidade e transparência
- Operações de tratamento verificáveis
- Transparência com titulares sobre como dados são usados
- Documentação acessível (política de privacidade clara)
- Mecanismos de auditoria independente

### 7. Respeito pela privacidade do usuário
- Manter o titular no centro das decisões
- Oferecer controles granulares ao usuário
- Interface clara para exercício de direitos
- Consentimento informado e genuíno

---

## Privacy by Default

Aplicação prática do princípio 2:
- **Coleta mínima**: solicitar apenas dados essenciais para a finalidade
- **Acesso mínimo**: restringir a quem estritamente precisa
- **Retenção mínima**: excluir quando não mais necessários
- **Compartilhamento desabilitado**: não compartilhar por padrão
- **Configurações restritivas**: opções mais privadas selecionadas por padrão

---

## Ciclo de Vida dos Dados e Ativos

### Fases do Ciclo de Vida (conforme Guia de Boas Práticas):
1. **Coleta**: obtenção dos dados do titular
2. **Retenção**: armazenamento em sistemas e bases
3. **Processamento**: operações realizadas com os dados
4. **Compartilhamento**: transmissão a terceiros internos/externos
5. **Eliminação**: destruição segura dos dados

### Ativos Organizacionais envolvidos:
- **Pessoas**: colaboradores, terceirizados, fornecedores
- **Processos**: fluxos de trabalho, procedimentos operacionais
- **Tecnologia**: sistemas, servidores, redes, dispositivos
- **Ambiente físico**: salas, data centers, armários, documentos

---

## Controles Prioritários por Fase

### Coleta:
- Consentimento ou base legal definida antes da coleta
- Minimização dos dados coletados
- Informação clara ao titular (aviso de privacidade)
- Registro da base legal

### Retenção:
- Criptografia em repouso
- Controle de acesso granular
- Classificação por sensibilidade
- Prazo de retenção definido

### Processamento:
- Logs de todas as operações
- Pseudonimização quando possível
- Validação de autorização
- Proteção contra erros

### Compartilhamento:
- Verificação de base legal para compartilhamento
- Criptografia em trânsito
- Contratos com terceiros (Art. 39)
- Registro do compartilhamento

### Eliminação:
- Destruição segura (overwrite, destruição física)
- Confirmação de eliminação documentada
- Eliminação de backups no prazo definido
- Notificação a terceiros que receberam os dados
