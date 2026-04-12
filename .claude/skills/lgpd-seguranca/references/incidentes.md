# Incidentes de Segurança - LGPD

## Sumário
1. [Fundamentação Legal (Art. 48)](#fundamentação-legal)
2. [Resolução CD/ANPD nº 15/2024](#resolução-152024)
3. [Quando Comunicar](#quando-comunicar)
4. [Prazos de Comunicação](#prazos)
5. [Conteúdo da Comunicação](#conteúdo)
6. [Plano de Resposta a Incidentes (PRI)](#plano-de-resposta)
7. [Modelo de Comunicação à ANPD](#modelo-comunicação)

---

## Fundamentação Legal (Art. 48)

**Art. 48 da LGPD**: "O controlador deverá comunicar à autoridade nacional e ao titular a
ocorrência de incidente de segurança que possa acarretar risco ou dano relevante aos titulares."

**Art. 48, § 1º**: A comunicação deve conter, no mínimo:
- I - Descrição da natureza dos dados pessoais afetados
- II - Informações sobre os titulares envolvidos
- III - Indicação das medidas técnicas e de segurança utilizadas (observado segredo comercial)
- IV - Riscos relacionados ao incidente
- V - Razões da eventual demora na comunicação (se não foi imediata)
- VI - Medidas adotadas ou a serem adotadas para reverter/mitigar efeitos

**Art. 48, § 2º**: A ANPD verificará a gravidade do incidente e poderá determinar medidas
adicionais como ampla divulgação do fato em meios de comunicação e medidas para reverter
ou mitigar os efeitos.

**Art. 48, § 3º**: No juízo de gravidade do incidente, será avaliada a comprovação de medidas
técnicas adequadas que tornem os dados ininteligíveis (como criptografia).

---

## Resolução CD/ANPD nº 15/2024

Publicada em 24 de abril de 2024, regulamenta a comunicação de incidentes de segurança:

### Definição de Incidente:
Qualquer evento adverso confirmado, relacionado à violação na segurança de dados pessoais,
que comprometa a confidencialidade, integridade ou disponibilidade dos dados.

### Avaliação de Risco:
O controlador deve desenvolver **metodologia própria** para avaliar se o incidente pode
causar risco ou dano relevante aos titulares, considerando:
- Natureza e categoria dos dados afetados
- Volume de dados envolvidos
- Se inclui dados sensíveis, de menores ou financeiros
- Facilidade de identificação dos titulares
- Consequências concretas e prováveis para os titulares

### Obrigações do Controlador:
1. Avaliar internamente o incidente
2. Comunicar à ANPD se houver risco ou dano relevante
3. Comunicar aos titulares afetados
4. Documentar todas as etapas
5. Adotar medidas para mitigar/reverter efeitos

---

## Quando Comunicar

### Incidentes que EXIGEM comunicação (risco/dano relevante):
- Envolvem **dados sensíveis** (origem racial, saúde, biometria, etc.)
- Envolvem **dados de crianças e adolescentes**
- Envolvem **dados financeiros** (contas, cartões, transações)
- Envolvem **dados de autenticação** em sistemas (senhas, tokens)
- Envolvem **dados protegidos por sigilo** (profissional, bancário, fiscal)
- Tratamento em **larga escala** (grande volume de titulares)
- Podem causar **dano patrimonial** (fraude, roubo de identidade)
- Podem causar **dano moral** (discriminação, exposição vexatória)
- Podem afetar **exercício de direitos** dos titulares

### Incidentes que podem NÃO exigir comunicação:
- Dados adequadamente criptografados e chaves não comprometidas
- Dados anonimizados (não são mais dados pessoais)
- Incidente contido antes de qualquer acesso aos dados
- Dados já publicamente disponíveis

### Na dúvida: COMUNIQUE. É melhor comunicar um incidente que se revele menor
do que deixar de comunicar um incidente relevante.

---

## Prazos de Comunicação

### À ANPD:
- **3 dias úteis** a partir do conhecimento do incidente
- Contados da data em que o controlador tomou conhecimento
- Comunicação deve ser feita pelo portal da ANPD

### Aos Titulares:
- **Prazo razoável** (simultaneamente ou logo após comunicação à ANPD)
- Meios adequados para alcançar os titulares afetados
- Comunicação individual quando possível
- Comunicação pública se inviável a individual

### Comunicação complementar:
Se informações não estiverem disponíveis nos 3 dias úteis, enviar comunicação
preliminar e complementar posteriormente.

---

## Conteúdo da Comunicação

### À ANPD (conteúdo mínimo):
1. Descrição da natureza dos dados pessoais afetados
2. Informações sobre os titulares envolvidos (quantidade, categorias)
3. Medidas técnicas e de segurança utilizadas para proteção
4. Riscos relacionados ao incidente com identificação dos possíveis impactos
5. Motivos da demora, se a comunicação não foi imediata
6. Medidas adotadas ou que serão adotadas para reverter/mitigar o prejuízo
7. Data do conhecimento do incidente
8. Dados do encarregado (nome e contato)

### Aos Titulares (conteúdo mínimo):
1. Descrição clara e acessível do incidente
2. Natureza dos dados pessoais afetados
3. Medidas adotadas para mitigar efeitos
4. Recomendações de ações que o titular pode tomar
5. Dados de contato do encarregado
6. Como obter mais informações

---

## Plano de Resposta a Incidentes (PRI)

### Fase 1: Preparação
- Equipe de resposta definida (CSIRT/equipe multidisciplinar)
- Papéis e responsabilidades documentados
- Canais de comunicação interna estabelecidos
- Ferramentas e recursos disponíveis
- Treinamentos e simulações periódicas (mínimo anual)

### Fase 2: Detecção e Análise
- Monitoramento contínuo (SIEM, IDS/IPS, DLP)
- Classificação da severidade do incidente
- Avaliação inicial: dados afetados, titulares impactados
- Decisão sobre comunicação à ANPD e titulares
- Registro detalhado (timestamp, evidências, decisões)

### Fase 3: Contenção
- Contenção imediata (isolar sistemas, bloquear acessos)
- Preservação de evidências para investigação
- Contenção de curto prazo (workarounds temporários)
- Contenção de longo prazo (correções permanentes)

### Fase 4: Erradicação
- Identificar e eliminar causa raiz
- Remover malware, contas comprometidas, vulnerabilidades
- Aplicar patches e correções
- Validar que a ameaça foi eliminada

### Fase 5: Recuperação
- Restaurar sistemas e dados a partir de backups íntegros
- Monitoramento intensificado pós-recuperação
- Validação de integridade dos dados restaurados
- Retorno gradual à operação normal

### Fase 6: Comunicação
- Comunicar ANPD em até 3 dias úteis (se aplicável)
- Comunicar titulares afetados
- Comunicar outras autoridades se necessário (Polícia, MP)
- Comunicação interna à alta gestão

### Fase 7: Lições Aprendidas
- Análise post-mortem documentada
- Identificação de melhorias necessárias
- Atualização de políticas e procedimentos
- Treinamento adicional se necessário
- Atualização do próprio PRI

---

## Modelo de Comunicação à ANPD

```
COMUNICAÇÃO DE INCIDENTE DE SEGURANÇA COM DADOS PESSOAIS

1. IDENTIFICAÇÃO DO CONTROLADOR
   - Razão Social: [nome]
   - CNPJ: [número]
   - Encarregado: [nome, e-mail, telefone]

2. DATA DO INCIDENTE
   - Data/hora da ocorrência (estimada): [data/hora]
   - Data/hora do conhecimento: [data/hora]
   - Data desta comunicação: [data]

3. DESCRIÇÃO DO INCIDENTE
   - Natureza: [acesso não autorizado/vazamento/perda/alteração/outro]
   - Descrição: [detalhamento do que ocorreu]
   - Causa provável: [se identificada]

4. DADOS PESSOAIS AFETADOS
   - Tipos de dados: [nome, CPF, e-mail, dados sensíveis, etc.]
   - Categorias de titulares: [clientes, funcionários, fornecedores, etc.]
   - Quantidade estimada de titulares: [número]
   - Dados sensíveis envolvidos: [sim/não - quais]

5. MEDIDAS DE SEGURANÇA PRÉ-EXISTENTES
   - [Listar medidas que estavam em vigor]
   - Dados estavam criptografados: [sim/não]

6. RISCOS IDENTIFICADOS
   - Possíveis impactos aos titulares: [fraude, discriminação, dano moral, etc.]
   - Probabilidade de materialização: [baixa/média/alta]

7. MEDIDAS ADOTADAS
   - Contenção: [medidas tomadas para conter o incidente]
   - Mitigação: [medidas para reduzir impacto aos titulares]
   - Comunicação aos titulares: [sim/não - quando/como]

8. MOTIVO DE EVENTUAL DEMORA
   - [Se comunicação não foi imediata, justificar]

[Local], [Data]
[Assinatura do Representante Legal ou Encarregado]
```
