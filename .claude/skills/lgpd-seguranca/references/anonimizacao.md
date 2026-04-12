# Anonimização e Pseudonimização

## Sumário
1. [Definições Legais](#definições)
2. [Diferenças Fundamentais](#diferenças)
3. [Técnicas de Anonimização](#técnicas-anonimização)
4. [Técnicas de Pseudonimização](#técnicas-pseudonimização)
5. [Riscos de Reidentificação](#riscos)
6. [Critérios de Escolha](#critérios-de-escolha)
7. [Recomendações Práticas](#recomendações)

---

## Definições Legais

### Dado Anonimizado (Art. 12):
"Dado relativo a titular que não possa ser identificado, considerando a utilização de meios
técnicos razoáveis e disponíveis na ocasião de seu tratamento."

O dado anonimizado **não é considerado dado pessoal** e portanto não se aplica a LGPD,
**salvo** quando o processo de anonimização puder ser revertido (Art. 12, caput).

### Pseudonimização (Art. 13, § 4º):
"Tratamento por meio do qual um dado perde a possibilidade de associação, direta ou
indireta, a um indivíduo, senão pelo uso de informação adicional mantida separadamente
pelo controlador em ambiente controlado e seguro."

Dado pseudonimizado **ainda é dado pessoal** e continua regulado pela LGPD.

---

## Diferenças Fundamentais

| Aspecto | Anonimização | Pseudonimização |
|---------|-------------|-----------------|
| **Reversibilidade** | Irreversível | Reversível (com chave) |
| **Status legal** | Não é dado pessoal | Continua sendo dado pessoal |
| **Aplicação da LGPD** | Não se aplica | Aplica-se integralmente |
| **Proteção ao titular** | Máxima (identidade eliminada) | Alta (identidade oculta, não eliminada) |
| **Utilidade dos dados** | Reduzida (dados genéricos) | Preservada (dados individualizáveis) |
| **Risco residual** | Risco de reidentificação | Risco se chave comprometida |
| **Uso recomendado** | Relatórios, estatísticas, pesquisa | Tratamento cotidiano, ambientes de teste |

---

## Técnicas de Anonimização

### Supressão:
- Remoção completa de identificadores diretos (nome, CPF, RG)
- Remove campos inteiros ou registros específicos
- Simples mas pode reduzir utilidade dos dados

### Generalização:
- Substituir valores específicos por faixas (idade 32 → 30-40)
- Reduzir granularidade geográfica (endereço → cidade → estado)
- Preserva padrões estatísticos, reduz identificação

### Perturbação/Ruído:
- Adicionar ruído aleatório aos dados numéricos
- Permutação de valores entre registros
- Arredondamento de valores
- Preserva distribuição estatística, dificulta identificação

### K-Anonimato:
- Garantir que cada registro é indistinguível de pelo menos k-1 outros
- Cada combinação de quasi-identificadores aparece pelo menos k vezes
- Proteção contra ataques por ligação

### L-Diversidade:
- Extensão do k-anonimato
- Cada grupo de k registros tem pelo menos l valores distintos para atributos sensíveis
- Protege contra ataques por homogeneidade

### Privacidade Diferencial:
- Técnica matemática que adiciona ruído calibrado
- Garante que a presença/ausência de um indivíduo não afeta significativamente o resultado
- Padrão ouro para anonimização estatística

---

## Técnicas de Pseudonimização

### Criptografia:
- Substituir identificadores por valores criptografados
- Chave mantida separadamente em ambiente seguro
- Reversível apenas por quem possui a chave
- Exemplo: CPF criptografado com AES-256

### Tokenização:
- Substituir dados sensíveis por tokens (valores aleatórios)
- Mapeamento token ↔ dado original em vault seguro
- Comum para dados de cartão de crédito (PCI DSS)
- Token não possui relação matemática com o dado original

### Hashing com Salt:
- Aplicar função hash ao identificador com salt aleatório
- Resultado é determinístico (mesmo input = mesmo output)
- Útil para vinculação de registros sem expor identidade
- Irreversível sem rainbow tables (com salt adequado)

### Substituição por Identificador Artificial:
- Criar ID único artificial para cada titular
- Mapear ID artificial ↔ identidade real em sistema separado
- Acessos ao mapeamento controlados e auditados

---

## Riscos de Reidentificação

### Ataques de Linkage:
- Cruzar dados anonimizados com bases externas públicas
- Combinar quasi-identificadores (data nascimento + CEP + gênero)
- Risco aumenta com disponibilidade de dados abertos

### Ataques por Inferência:
- Deduzir informações sobre indivíduos a partir de dados agregados
- Grupos muito pequenos permitem identificação
- Dados sensíveis podem ser inferidos de dados não-sensíveis

### Risco de IA/ML:
- Algoritmos de aprendizado de máquina podem reidentificar dados
- Técnicas avançadas superam anonimização básica
- Risco cresce com avanço tecnológico

### Fatores que aumentam o risco:
- Datasets pequenos ou muito específicos
- Muitas variáveis quasi-identificadoras
- Disponibilidade de dados auxiliares para cruzamento
- Dados de populações pequenas ou muito específicas

### Mitigação:
- Avaliação periódica de risco de reidentificação
- Considerar "meios técnicos razoáveis" conforme Art. 12
- Documentar avaliação e justificativa
- Revisar quando novas técnicas surgirem

---

## Critérios de Escolha

### Escolher ANONIMIZAÇÃO quando:
- Dados serão usados apenas para estatísticas ou pesquisa
- Não há necessidade de vincular dados a indivíduos no futuro
- Deseja-se eliminar obrigações da LGPD sobre aqueles dados
- Dados serão publicados ou amplamente compartilhados
- Obrigatório quando não há necessidade de guardar associação com titular

### Escolher PSEUDONIMIZAÇÃO quando:
- Operações cotidianas exigem vinculação futura com titular
- Ambiente de desenvolvimento/teste necessita dados realistas
- Titular pode exercer direitos (acesso, correção, exclusão)
- Dados de pesquisa em saúde pública (Art. 13)
- Necessidade de reidentificação controlada
- Proteção de identidade de denunciante (Decreto 10.153/2019)
- Proteção de usuário de serviço público (Decreto 9.492/2018)

---

## Recomendações Práticas

1. **Elencar processos** que realizam tratamento e identificar dados sem titulares vinculados
2. **Analisar ciclo de vida** para mitigar riscos de violação de dados não mais em uso
3. **Avaliar risco de identificação** considerando volume, tecnologias de análise e significância
4. **Optar por anonimização** quando não há necessidade de guardar associação
5. **Optar por pseudonimização** quando precisar manter vinculação controlada
6. **Definir plano de comunicação** para incidentes de violação de dados
7. **Documentar violações** e incidentes periodicamente para análise de riscos
8. **Promover conscientização** contínua sobre proteção de dados
