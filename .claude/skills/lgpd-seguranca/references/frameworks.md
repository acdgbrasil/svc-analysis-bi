# Frameworks e Normas de Segurança da Informação

## Sumário
1. [Comparativo Geral](#comparativo)
2. [ISO/IEC 27001:2022](#iso-27001)
3. [ISO/IEC 27701:2025](#iso-27701)
4. [ISO/IEC 27005:2019 - Gestão de Riscos](#iso-27005)
5. [ISO/IEC 29134:2017 - Avaliação de Impacto](#iso-29134)
6. [NIST Cybersecurity Framework (CSF)](#nist-csf)
7. [NIST Privacy Framework (PF)](#nist-pf)
8. [e-Ping - Governo Eletrônico](#e-ping)
9. [ABNT NBR ISO/IEC 27002:2013](#iso-27002)
10. [Normativos GSI/PR](#gsi)
11. [Resoluções CONARQ](#conarq)
12. [Guia de Escolha](#guia-de-escolha)

---

## Comparativo Geral

| Framework | Tipo | Certificável | Foco | Melhor para |
|-----------|------|-------------|------|-------------|
| ISO 27001 | Prescritivo | Sim | Segurança da informação | Organizações que buscam certificação |
| ISO 27701 | Prescritivo | Sim | Privacidade (extensão 27001) | Compliance LGPD/GDPR com certificação |
| NIST CSF | Adaptável | Não | Cybersecurity | Organizações que preferem flexibilidade |
| NIST PF | Adaptável | Não | Privacidade | Complemento ao CSF para privacidade |
| ISO 27005 | Metodológico | Não | Gestão de riscos | RIPD e avaliação de riscos |
| ISO 29134 | Metodológico | Não | Avaliação de impacto | Elaboração de RIPD |

---

## ISO/IEC 27001:2022

### O que é:
Norma internacional para **Sistemas de Gestão da Segurança da Informação (SGSI)**.
Estabelece requisitos para implementar, manter e melhorar continuamente um SGSI.

### Estrutura:
- 93 controles organizados em 4 categorias (Anexo A):
  - **Organizacionais** (37 controles): políticas, papéis, classificação
  - **Pessoais** (8 controles): seleção, conscientização, disciplinar
  - **Físicos** (14 controles): perímetros, equipamentos, mídias
  - **Tecnológicos** (34 controles): acesso, criptografia, logs, desenvolvimento

### Relevância para LGPD:
- Base para atender Art. 46 (medidas de segurança)
- Demonstra responsabilização e prestação de contas (Art. 6º, X)
- Certificação como evidência de conformidade
- Amplamente reconhecida pela ANPD como referência

### No Brasil:
- ABNT NBR ISO/IEC 27001:2022 (versão brasileira)
- Certificação disponível via organismos acreditados pelo INMETRO

---

## ISO/IEC 27701:2025

### O que é:
Framework **independente** (desde 2025, não mais apenas extensão da 27001) para **gestão de
privacidade da informação**. Integra cybersecurity + privacy management.

### Novidades da versão 2025:
- Framework autônomo (pode ser implementado independentemente da 27001)
- Controles específicos de privacidade expandidos
- **Anexo D**: Mapeamento detalhado com GDPR (aplicável à LGPD por similaridade)
- Abordagem integrada de segurança e privacidade

### Relevância para LGPD:
- Mapeamento direto com princípios e obrigações da LGPD
- Como LGPD é inspirada na GDPR, o Anexo D serve como guia prático
- Certificação como evidência de conformidade com privacidade
- Cobre: consentimento, direitos do titular, RIPD, incidentes, transferência internacional

### Implementação:
- Define papéis de controlador (PII Controller) e operador (PII Processor)
- Controles para gestão do consentimento
- Controles para exercício de direitos dos titulares
- Controles para comunicação de incidentes
- Controles para avaliação de impacto à privacidade

---

## ISO/IEC 27005:2019 - Gestão de Riscos

### O que é:
Metodologia para **gestão de riscos de segurança da informação**.

### Processo:
1. Estabelecimento do contexto
2. Identificação de riscos
3. Análise de riscos (probabilidade × impacto)
4. Avaliação de riscos (priorização)
5. Tratamento de riscos (evitar, mitigar, transferir, aceitar)
6. Monitoramento e revisão
7. Comunicação e consulta

### Relevância para LGPD:
- Essencial para elaboração do RIPD (Art. 38)
- Base para avaliação de riscos de incidentes (Art. 48)
- Suporta princípio da prevenção (Art. 6º, VIII)

### No Brasil:
- ABNT NBR ISO/IEC 27005:2019

---

## ISO/IEC 29134:2017 - Avaliação de Impacto

### O que é:
Diretrizes para **avaliação de impacto à privacidade** (Privacy Impact Assessment - PIA).

### Relevância para LGPD:
- Referência direta para elaboração do RIPD
- Seção 6.4.4: lista de riscos de privacidade (usada no Guia de Boas Práticas do Gov.br)
- Metodologia estruturada para identificação e avaliação de riscos à privacidade
- Base para a Tabela de Riscos do RIPD

---

## NIST Cybersecurity Framework (CSF)

### O que é:
Framework **adaptável e baseado em resultados** para gestão de riscos cibernéticos,
desenvolvido pelo National Institute of Standards and Technology (EUA).

### 5 Funções Core:
1. **Identificar (Identify)**: inventário de ativos, governança, avaliação de risco
2. **Proteger (Protect)**: controle de acesso, treinamento, manutenção
3. **Detectar (Detect)**: monitoramento, detecção de anomalias
4. **Responder (Respond)**: plano de resposta, comunicação, mitigação
5. **Recuperar (Recover)**: plano de recuperação, melhorias, comunicação

### Vantagens:
- Flexível e adaptável a qualquer porte de organização
- Não prescritivo (define "o quê", não "como")
- Gratuito e amplamente documentado
- Versões em português disponíveis

### Relevância para LGPD:
- Complementa requisitos técnicos do Art. 46
- Estrutura para gestão de incidentes (Art. 48)
- Pode ser combinado com ISO 27001

---

## NIST Privacy Framework (PF)

### O que é:
Framework focado em **privacidade**, complementar ao NIST CSF.

### Funções:
1. **Identificar-P**: inventário de dados, avaliação de riscos de privacidade
2. **Governar-P**: políticas, papéis, conscientização
3. **Controlar-P**: gestão de dados, consentimento, minimização
4. **Comunicar-P**: transparência, políticas de privacidade
5. **Proteger-P**: medidas técnicas de proteção de privacidade

### Relevância para LGPD:
- Complementa NIST CSF para aspectos de privacidade
- Alinha-se com princípios da LGPD
- Abordagem flexível adaptável ao contexto brasileiro

---

## e-Ping - Governo Eletrônico

### O que é:
**Padrões de Interoperabilidade de Governo Eletrônico**, arquitetura de referência que
define políticas e especificações técnicas para o governo federal brasileiro.

### Relevância para LGPD:
- Obrigatório para órgãos do governo federal
- Define padrões de segurança para interoperabilidade
- Inclui especificações para proteção de dados em serviços públicos digitais

---

## ABNT NBR ISO/IEC 27002:2013

### O que é:
**Código de prática** para controles de segurança da informação.
Complementa a ISO 27001 com orientações de implementação detalhadas.

### Relevância:
- Guia prático para implementação dos controles da 27001
- Referência para medidas técnicas do Art. 46 da LGPD
- Detalhamento de controles de acesso, criptografia, operações, etc.

---

## Normativos GSI/PR

### Gabinete de Segurança Institucional da Presidência da República:
- Emite normativos de segurança da informação para a Administração Pública Federal
- Instrução Normativa GSI/PR nº 1 (Política de Segurança da Informação)
- Normas Complementares (NC): detalham controles específicos

### Relevância para LGPD:
- Obrigatório para órgãos e entidades da APF
- Complementa as exigências técnicas da LGPD no setor público
- Define padrões mínimos de segurança

---

## Resoluções CONARQ

### Conselho Nacional de Arquivos:
- **Resolução nº 25/2007**: Modelo de Requisitos para SIGAD
- **Resolução nº 39/2014**: Diretrizes para repositórios digitais confiáveis

### Relevância para LGPD:
- Dados pessoais frequentemente estão em documentos arquivísticos
- LGPD e legislação arquivística devem ser interpretadas sistematicamente
- Retenção e eliminação de dados devem considerar prazos arquivísticos

---

## Guia de Escolha

### Para organizações do setor público brasileiro:
**Obrigatório**: e-Ping + Normativos GSI/PR + CONARQ
**Recomendado**: ISO 27001 + ISO 27701

### Para empresas de médio/grande porte:
**Recomendado**: ISO 27001:2022 + ISO 27701:2025 (certificação)
**Complementar**: NIST CSF + NIST PF (flexibilidade)

### Para startups e pequenas empresas:
**Recomendado**: NIST CSF (gratuito, flexível)
**À medida que cresce**: migrar para ISO 27001

### Para RIPD e avaliação de riscos:
**Primário**: ISO 27005 + ISO 29134
**Complementar**: NIST RMF

### Para demonstrar conformidade à ANPD:
**Mais forte**: Certificação ISO 27001 + ISO 27701
**Alternativa**: Programa documentado baseado em NIST CSF + PF
