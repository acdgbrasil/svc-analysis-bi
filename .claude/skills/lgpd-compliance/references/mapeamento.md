# Mapeamento de Dados e Inventário - LGPD

## Sumário
1. [Registro de Operações - ROPA (Art. 37)](#ropa)
2. [Processo de Mapeamento](#processo-de-mapeamento)
3. [Inventário de Dados Pessoais](#inventário)
4. [Gap Analysis](#gap-analysis)
5. [Modelo de ROPA](#modelo-ropa)
6. [Modelo de Inventário](#modelo-inventário)

---

## Registro de Operações - ROPA (Art. 37)

**Art. 37 da LGPD**: "O controlador e o operador devem manter registro das operações de
tratamento de dados pessoais que realizarem, especialmente quando baseado no legítimo interesse."

### Objetivo:
- Instalar reflexão sobre o uso responsável de dados pessoais
- Garantir accountability (prestação de contas - Art. 6º, X)
- Viabilizar auditorias internas e externas
- Estar preparado para solicitações da ANPD

### Conteúdo Mínimo Recomendado:
- Nome e descrição da operação de tratamento
- Base legal atribuída (Arts. 7º ou 11)
- Categorias de dados pessoais tratados
- Categorias de titulares
- Finalidade do tratamento
- Operadores envolvidos
- Informações sobre transferências internacionais
- Prazo de retenção dos dados
- Medidas de segurança aplicadas
- Responsável pelo registro

### Formato:
- Pode ser eletrônico ou físico
- Deve ser acessível para verificação pela ANPD
- Recomenda-se formato eletrônico e estruturado (planilha, sistema de gestão)

### Frequência de Atualização:
- Mínimo: anual
- Recomendado: trimestral ou quando houver mudanças significativas
- Obrigatório: quando novos tratamentos forem iniciados

---

## Processo de Mapeamento

### Etapa 1: Preparação
- Definir escopo do mapeamento (toda organização ou por área)
- Identificar stakeholders-chave por área/departamento
- Preparar questionários e templates
- Definir cronograma

### Etapa 2: Identificação de Fontes
- Onde dados pessoais são coletados (formulários, sites, apps, presencial)
- Como dados entram na organização (canais de entrada)
- Quais sistemas/bases armazenam dados pessoais
- Quais fornecedores/parceiros recebem dados

### Etapa 3: Catalogação de Dados
- Tipos de dados pessoais (nome, CPF, e-mail, endereço, etc.)
- Tipos de dados sensíveis (saúde, biometria, origem racial, etc.)
- Volume aproximado por categoria
- Classificação de sensibilidade (baixa, média, alta, crítica)

### Etapa 4: Mapeamento de Fluxos
- Fluxo interno: como dados circulam entre áreas/sistemas
- Fluxo externo: transferências para terceiros
- Transferências internacionais (se aplicável)
- Diagramas de fluxo de dados (DFD)

### Etapa 5: Identificação de Titulares
- Categorias: clientes, funcionários, candidatos, fornecedores, visitantes, etc.
- Número aproximado por categoria
- Titulares vulneráveis (crianças, idosos, etc.)

### Etapa 6: Base Legal e Finalidade
- Para cada tratamento: qual base legal (Art. 7º ou 11)
- Para cada tratamento: qual finalidade específica
- Verificar se a base legal é adequada
- Documentar justificativa

### Etapa 7: Retenção e Descarte
- Prazo de retenção para cada tipo de dado
- Fundamentação legal para o prazo (obrigação legal, necessidade operacional)
- Procedimento de descarte definido
- Responsável pelo descarte

### Etapa 8: Medidas de Segurança
- Controles técnicos aplicados a cada tratamento
- Controles administrativos aplicados
- Gaps identificados
- Plano de remediação

---

## Inventário de Dados Pessoais

### Componentes:
- Lista centralizada de todos os dados pessoais tratados
- Base legal para cada tratamento
- Categorias de titulares
- Finalidades específicas
- Prazos de retenção
- Localização de armazenamento (sistema, servidor, nuvem)
- Responsáveis pelo tratamento
- Operadores envolvidos
- Status de conformidade
- Data da última revisão

### Atualização:
- Mínimo anual, recomendado trimestral
- Obrigatório quando:
  - Novo tratamento iniciado
  - Mudança significativa em tratamento existente
  - Incidente de segurança ocorrer
  - Regulamentação nova da ANPD

---

## Gap Analysis

### O que é:
Análise comparativa entre o estado atual de conformidade e o estado desejado (plena
adequação à LGPD), identificando lacunas que precisam ser endereçadas.

### Etapa 1: Baseline Atual
- Avaliar conformidade atual em cada área
- Documentar práticas existentes
- Identificar controles já implementados
- Entrevistar áreas-chave

### Etapa 2: Estado Desejado
- Conformidade total com a LGPD
- Alinhamento com resoluções da ANPD
- Atendimento a frameworks escolhidos (ISO, NIST)
- Melhores práticas do setor

### Etapa 3: Identificação de Lacunas
Avaliar em cada dimensão:

**Governança:**
- [ ] DPO/Encarregado nomeado e publicizado?
- [ ] Programa de governança documentado (Art. 50)?
- [ ] Comitê de privacidade instituído?
- [ ] Alta gestão comprometida?

**Bases Legais:**
- [ ] Todas as operações têm base legal identificada?
- [ ] Consentimento obtido corretamente quando necessário?
- [ ] Legítimo interesse documentado com teste de balanceamento?

**Direitos dos Titulares:**
- [ ] Canal do titular implementado?
- [ ] Processos para atender Art. 18 definidos?
- [ ] Prazos de resposta documentados?

**Segurança:**
- [ ] Medidas técnicas implementadas (Art. 46)?
- [ ] Plano de resposta a incidentes (Art. 48)?
- [ ] Controles de acesso adequados?

**Documentação:**
- [ ] ROPA atualizado (Art. 37)?
- [ ] RIPDs elaborados quando necessário (Art. 38)?
- [ ] Políticas internas documentadas?

**Contratos:**
- [ ] Contratos com operadores revisados (Art. 39)?
- [ ] Cláusulas de proteção de dados incluídas?

**Treinamento:**
- [ ] Programa de treinamento implementado?
- [ ] Registro de participação?

**Transferência Internacional:**
- [ ] Mecanismos adequados para transferência (Arts. 33-36)?

### Etapa 4: Plano de Remediação
- Priorizar lacunas por risco (alto/médio/baixo)
- Estimar esforço e recursos necessários
- Definir prazos realistas
- Designar responsáveis
- Estabelecer marcos de acompanhamento

---

## Modelo de ROPA

```
REGISTRO DE OPERAÇÕES DE TRATAMENTO DE DADOS PESSOAIS

| Campo | Conteúdo |
|-------|----------|
| ID da Operação | [ROPA-001] |
| Nome da Operação | [Ex: Cadastro de Clientes] |
| Área Responsável | [Ex: Comercial] |
| Descrição | [Breve descrição do tratamento] |
| Base Legal | [Art. 7º, I - Consentimento] |
| Finalidade | [Ex: Prestação de serviços contratados] |
| Categorias de Dados | [Nome, CPF, e-mail, telefone] |
| Dados Sensíveis | [Sim/Não - quais] |
| Categorias de Titulares | [Clientes PF] |
| Volume Estimado | [10.000 registros] |
| Fonte dos Dados | [Formulário web] |
| Operadores | [Empresa X - hosting] |
| Compartilhamento | [Financeiro interno, Contabilidade] |
| Transferência Internacional | [Sim/Não - país, mecanismo] |
| Prazo de Retenção | [5 anos após encerramento do contrato] |
| Medidas de Segurança | [Criptografia, controle de acesso, backup] |
| Data de Criação | [dd/mm/aaaa] |
| Última Atualização | [dd/mm/aaaa] |
| Responsável | [Nome] |
```

---

## Modelo de Inventário

```
INVENTÁRIO DE DADOS PESSOAIS

Organização: [Nome]
Data: [dd/mm/aaaa]
Responsável: [Nome do DPO]

| Dado | Tipo | Sensível | Finalidade | Base Legal | Sistema | Retenção | Conformidade |
|------|------|----------|-----------|-----------|---------|----------|-------------|
| Nome | Cadastral | Não | Identificação | Contrato | CRM | 5 anos | Conforme |
| CPF | Cadastral | Não | Obrigação fiscal | Obrig. legal | ERP | 10 anos | Conforme |
| Biometria | Biométrico | Sim | Controle acesso | Prev. fraude | Ponto | 1 ano | Em adequação |
| E-mail | Contato | Não | Comunicação | Consentimento | CRM | Até revogação | Conforme |
```
