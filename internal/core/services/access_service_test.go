package services_test

import (
	"context"
	"errors"

	"testing"

	"github.com/paulochiaradia/gatekeeper/internal/core/domain"
	"github.com/paulochiaradia/gatekeeper/internal/core/services"
	"github.com/stretchr/testify/assert"
)

// ==========================================
// 1. O MOCK DO BANCO DE DADOS EM MEMÓRIA
// ==========================================

// MockRepository simula o banco implementando as duas interfaces (UserRepository e AccessLogRepository)
type MockRepository struct {
	tabelaUsuarios map[string]*domain.User
	logsGravados   []*domain.AccessLog
}

// GetUserByUID simula a busca (SELECT)
func (m *MockRepository) GetUserByUID(_ context.Context, uid string) (*domain.User, error) {
	if user, existe := m.tabelaUsuarios[uid]; existe {
		return user, nil
	}
	return nil, errors.New("usuário não encontrado")
}

// SaveLog simula a gravação da auditoria (INSERT)
func (m *MockRepository) SaveLog(_ context.Context, log *domain.AccessLog) error {
	m.logsGravados = append(m.logsGravados, log)
	return nil
}

// CreateUser é exigido pela interface, mas não usaremos nesse teste
func (m *MockRepository) CreateUser(_ context.Context, user *domain.User) error { return nil }

// ==========================================
// 2. A SUÍTE DE TESTES (TABLE-DRIVEN)
// ==========================================

func TestProcessAccessRequest(t *testing.T) {
	// Pré-carregamos o banco de mentira com dois usuários
	bancoFalso := &MockRepository{
		tabelaUsuarios: map[string]*domain.User{
			"UID-VALIDO":  {UID: "UID-VALIDO", Name: "Paulo", IsActive: true},
			"UID-INATIVO": {UID: "UID-INATIVO", Name: "Visitante", IsActive: false},
		},
		logsGravados: make([]*domain.AccessLog, 0),
	}

	// Instanciamos o serviço passando o banco falso
	accessService := services.NewAccessService(bancoFalso, bancoFalso)

	// Definimos todos os cenários possíveis
	cenarios := []struct {
		nomeDoCenario   string
		uidTestado      string
		esperaConcedido bool
		esperaErro      string
		esperaStatusLog string
	}{
		{
			nomeDoCenario:   "Caminho Feliz - Usuário Válido e Ativo",
			uidTestado:      "UID-VALIDO",
			esperaConcedido: true,
			esperaErro:      "",
			esperaStatusLog: "GRANTED",
		},
		{
			nomeDoCenario:   "Falha - Usuário Inativo",
			uidTestado:      "UID-INATIVO",
			esperaConcedido: false,
			esperaErro:      "inativo",
			esperaStatusLog: "DENIED_INACTIVE",
		},
		{
			nomeDoCenario:   "Falha - Crachá Desconhecido",
			uidTestado:      "UID-DESCONHECIDO",
			esperaConcedido: false,
			esperaErro:      "não encontrado",
			esperaStatusLog: "",
		},
	}

	// O motor que roda os cenários um por um
	for _, cenario := range cenarios {
		t.Run(cenario.nomeDoCenario, func(t *testing.T) {
			// ISOLAMENTO: Limpa os logs do banco de mentira antes de CADA cenário
			bancoFalso.logsGravados = make([]*domain.AccessLog, 0)

			// Executa a função real (agora passando o context.Background() para o teste)
			concedido, err := accessService.ProcessAccessRequest(context.Background(), cenario.uidTestado)

			// Validações
			assert.Equal(t, cenario.esperaConcedido, concedido)

			if cenario.esperaErro != "" {
				// Usamos Contains para flexibilidade: basta a frase conter "não encontrado"
				assert.Contains(t, err.Error(), cenario.esperaErro)
			} else {
				assert.NoError(t, err)
			}

			// Validação Condicional do Log
			if cenario.esperaStatusLog != "" {
				if assert.GreaterOrEqual(t, len(bancoFalso.logsGravados), 1, "Um log deveria ter sido gravado") {
					logGerado := bancoFalso.logsGravados[0]
					assert.Equal(t, cenario.esperaStatusLog, logGerado.Status)
					assert.Equal(t, cenario.uidTestado, logGerado.UserUID)
				}
			} else {
				assert.Empty(t, bancoFalso.logsGravados, "Nenhum log deveria ter sido gravado neste cenário")
			}
		})
	}
}
