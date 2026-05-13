package services

import (
	"crypto/rand"
	"errors"
	"fmt"
	"time"

	"github.com/rafael-mingossi/diario-kids-api/internal/dto"
	"github.com/rafael-mingossi/diario-kids-api/internal/mappers"
	"github.com/rafael-mingossi/diario-kids-api/internal/repository"
)

// interface publica
type AlunoService interface {
	CriarAluno(input dto.CriarAlunoInput) (*dto.AlunoResponse, error)
}

// interface privada
type alunoService struct {
	repo       repository.AlunoRepository
	salaRepo   repository.SalaRepository
	escolaRepo repository.EscolaRepository
}

// construtor devolve a interfaxe
func NewAlunoService(repo repository.AlunoRepository, salaRepo repository.SalaRepository, escolaRepo repository.EscolaRepository) AlunoService {
	return &alunoService{repo: repo, salaRepo: salaRepo, escolaRepo: escolaRepo}
}

// Sentinels de domínio: representam falhas esperadas de regra/entrada.
// O handler usa errors.Is para transformar esses casos em 422, sem tratar como 500.
var ErrDataNascimentoInvalida = errors.New("data_nascimento inválida")
var ErrDataNascimentoFutura = errors.New("data_nascimento não pode ser futura")
var ErrSalaNaoEncontrada = errors.New("sala informada não existe")
var ErrSalaNaoPertenceAEscola = errors.New("sala informada não pertence à escola")

const alfabetoMatricula = "ABCDEFGHJKLMNPQRSTUVWXYZ23456789"

// gerarMatricula cria o identificador de negócio do aluno no formato ALU-26-8H3K7Q.
// Regras do formato:
//   - ALU: prefixo fixo para alunos
//   - 26: ano curto (2026 -> 26)
//   - 8H3K7Q: 6 caracteres aleatórios em um alfabeto sem ambiguidade visual
//
// O alfabeto remove letras/números que costumam confundir na leitura manual:
//   - O e 0
//   - I e 1
//   - em geral evitamos símbolos também
func (s *alunoService) gerarMatricula() (string, error) {
	anoCurto := time.Now().Format("06")
	const tamanhoSufixo = 6
	randomBytes := make([]byte, tamanhoSufixo)
	_, err := rand.Read(randomBytes)
	if err != nil {
		return "", fmt.Errorf("erro ao gerar bytes aleatórios da matrícula: %w", err)
	}

	sufixo := make([]byte, tamanhoSufixo)
	for i, valor := range randomBytes {
		// Cada byte aleatório é transformado em um índice válido dentro do alfabeto permitido.
		// Isso nos dá uma matrícula curta, legível e com baixa chance de colisão.
		sufixo[i] = alfabetoMatricula[int(valor)%len(alfabetoMatricula)]
	}

	return fmt.Sprintf("ALU-%s-%s", anoCurto, string(sufixo)), nil
}

// service recebe o DTO de input, processa as regras de negocio e devolve o DTO de output
func (s *alunoService) CriarAluno(input dto.CriarAlunoInput) (*dto.AlunoResponse, error) {

	// O DTO traz a data como string porque esse formato é mais amigável para o frontend.
	// Aqui convertemos para time.Time, que é o tipo correto para persistir no model.
	// Parse vem ANTES da validação de data futura, porque enquanto a data ainda é string
	// nós não temos um valor temporal real para comparar com "hoje".
	dataNascimento, err := time.ParseInLocation("02/01/2006", input.DataNascimento, time.Local)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrDataNascimentoInvalida, err)
	}

	// Como nascimento é uma data sem horário, normalizamos "hoje" para 00:00.
	// Isso evita bugs de fuso/horário ao comparar apenas a parte da data.
	agora := time.Now().In(time.Local)
	hoje := time.Date(agora.Year(), agora.Month(), agora.Day(), 0, 0, 0, 0, agora.Location())
	if dataNascimento.After(hoje) {
		return nil, ErrDataNascimentoFutura
	}

	// EscolaID é obrigatório para manter o isolamento entre escolas/filiais.
	escola, err := s.escolaRepo.BuscarPorID(input.EscolaID)
	if err != nil {
		return nil, fmt.Errorf("erro ao verificar escola: %w", err)
	}
	if escola == nil {
		return nil, ErrEscolaNaoEncontrada
	}

	// SalaID é opcional no create inicial.
	// Só consultamos o banco se o cliente realmente informou uma sala.
	if input.SalaID != nil {
		sala, err := s.salaRepo.BuscarPorID(*input.SalaID)
		if err != nil {
			return nil, fmt.Errorf("erro ao verificar sala: %w", err)
		}
		if sala == nil {
			return nil, ErrSalaNaoEncontrada
		}
		// Sala e aluno precisam pertencer à mesma unidade.
		if sala.EscolaID != input.EscolaID {
			return nil, ErrSalaNaoPertenceAEscola
		}
	}

	// Com a data já validada e convertida, o mapper só monta o model sem regra extra.
	// O service é responsável por regras de negócio e validação, mas o mapper é burro por design.
	novoAluno := mappers.CriarAlunoInputToModel(input, dataNascimento)

	// A matrícula não vem do frontend: ela é gerada pelo backend no momento da criação.
	// Como o sufixo é aleatório, existe uma chance muito pequena de colisão.
	// Por isso tentamos algumas vezes antes de desistir.
	const maxTentativasMatricula = 5
	for tentativa := 1; tentativa <= maxTentativasMatricula; tentativa++ {
		matricula, err := s.gerarMatricula()
		if err != nil {
			return nil, fmt.Errorf("erro ao gerar matrícula: %w", err)
		}
		novoAluno.Matricula = matricula

		// Persiste o model já com matrícula preenchida.
		err = s.repo.Criar(&novoAluno)
		if err == nil {
			break
		}

		if !errors.Is(err, repository.ErrAlunoDuplicadoDB) {
			return nil, fmt.Errorf("erro ao criar aluno: %w", err)
		}

		// Colisão de matrícula: geramos outra e tentamos de novo.
		if tentativa == maxTentativasMatricula {
			return nil, fmt.Errorf("falha ao gerar matrícula única após %d tentativas: %w", maxTentativasMatricula, err)
		}
	}

	//converte model para DTO de resposta
	response := mappers.ModelToAlunoResponse(novoAluno)

	return &response, nil
}
