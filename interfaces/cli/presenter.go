package cli

import (
	"fmt"
	"go-expert-stress-test/domain"
	"strings"
	"time"
)

const (
	bold        = "\033[1m"
	colorReset  = "\033[0m"
	colorRed    = "\033[31m"
	colorGreen  = "\033[32m"
	colorYellow = "\033[33m"
	colorBlue   = "\033[34m"
	colorPurple = "\033[35m"
	colorCyan   = "\033[36m"
	colorGray   = "\033[37m"
)

type ReportPresenter struct{}

func NewReportPresenter() *ReportPresenter {
	return &ReportPresenter{}
}

// banerzinho pra fazer um fru-fru :)
func (p *ReportPresenter) displayBanner() {
	banner := `
█▀▀▀▀▀▀▀▀▀▀▀▀▀▀▀▀▀▀▀▀▀▀▀▀▀▀▀▀▀▀▀▀▀▀▀▀▀▀▀▀▀▀▀▀▀▀▀▀▀▀▀█
█  ╔═╗╔╦╗╦═╗╔═╗╔═╗╔═╗  ╔╦╗╔═╗╔═╗╔╦╗╔═╗╦═╗           █
█  ╚═╗ ║ ╠╦╝║╣ ╚═╗╚═╗   ║ ║╣ ╚═╗ ║ ║╣ ╠╦╝  ▄▄▄▄▄▄▄  █
█  ╚═╝ ╩ ╩╚═╚═╝╚═╝╚═╝   ╩ ╚═╝╚═╝ ╩ ╚═╝╩╚═  ▀▀▀▀▀▀▀  █
█▄▄▄▄▄▄▄▄▄▄▄▄▄▄▄▄▄▄▄▄▄▄▄▄▄▄▄▄▄▄▄▄▄▄▄▄▄▄▄▄▄▄▄▄▄▄▄▄▄▄▄█`
	fmt.Printf("%s%s%s\n", colorCyan, banner, colorReset)
}

// renderiza a barra de progresso de cada worker
func (p *ReportPresenter) DisplayProgressBar(current, total int) {
	width := 50
	progress := float64(current) / float64(total)
	completed := int(progress * float64(width))

	fmt.Printf("\r%s[", colorGray)
	for i := 0; i < width; i++ {
		if i < completed {
			fmt.Printf("%s█", colorGreen)
		} else {
			fmt.Printf("%s░", colorGray)
		}
	}
	fmt.Printf("%s] %d/%d (%.1f%%)%s",
		colorGray,
		current,
		total,
		progress*100,
		colorReset)
}

// obtem uma cor especifica para cada tipo de status code, outros codigos usan cor cinza como padrao
func (p *ReportPresenter) getStatusColor(status int) string {
	switch {
	case status >= 500:
		return colorRed
	case status >= 400:
		return colorYellow
	case status >= 300:
		return colorBlue
	case status >= 200:
		return colorGreen
	default:
		return colorGray
	}
}

// printa o resultados de cada http status code
func (p *ReportPresenter) displayStatusGraph(distrib map[int]int, total int) {
	fmt.Printf("\n\n%s%s= Distribuição de Status HTTP =%s\n", bold, colorPurple, colorReset)

	// encontra o maior valor para escalar o gráfico
	maxCount := 0
	for _, count := range distrib {
		if count > maxCount {
			maxCount = count
		}
	}
	maxWidth := 40

	// ordena e exibir as barras
	for status, count := range distrib {
		percentage := float64(count) / float64(total) * 100
		barWidth := int(float64(count) / float64(maxCount) * float64(maxWidth))

		statusColor := p.getStatusColor(status)
		bar := strings.Repeat("█", barWidth)

		fmt.Printf("%sHTTP %d %s[%s%s%s] %d (%.1f%%)\n",
			bold,
			status,
			colorGray,
			statusColor,
			bar,
			colorGray,
			count,
			percentage,
		)
	}
}

// imprime os resultados do teste
func (p *ReportPresenter) Present(report *domain.TestReport) {
	p.displayBanner()

	fmt.Printf("\n%s%s[RELATÓRIO DE TESTE DE CARGA]%s\n", bold, colorCyan, colorReset)
	fmt.Printf("\n%s▶ Métricas Gerais%s\n", colorPurple, colorReset)
	fmt.Printf("  • Duração Total: %s%v%s\n", colorGreen, report.TotalDuration.Round(time.Millisecond), colorReset)
	fmt.Printf("  • Total Requests: %s%d%s\n", colorGreen, report.TotalRequests, colorReset)
	fmt.Printf("  • Média por Request: %s%v%s\n", colorGreen, report.AverageDuration.Round(time.Millisecond), colorReset)

	successRate := float64(report.SuccessRequests) / float64(report.TotalRequests) * 100
	fmt.Printf("\n%s▶ Taxa de Sucesso%s\n", colorPurple, colorReset)
	fmt.Printf("  • Requests OK (2xx): %s%d (%.1f%%)%s\n",
		colorGreen,
		report.SuccessRequests,
		successRate,
		colorReset)

	if report.ErrorCount > 0 {
		errorRate := float64(report.ErrorCount) / float64(report.TotalRequests) * 100
		fmt.Printf("  • Erros: %s%d (%.1f%%)%s\n",
			colorRed,
			report.ErrorCount,
			errorRate,
			colorReset)
	}

	p.displayStatusGraph(report.StatusDistrib, report.TotalRequests)

	// sumario final do test de carga
	fmt.Printf("\n%s%s[SUMÁRIO]%s\n", bold, colorCyan, colorReset)
	if successRate >= 95 {
		fmt.Printf("%s✓ Sistema respondeu bem ao teste de carga%s\n", colorGreen, colorReset)
	} else if successRate >= 80 {
		fmt.Printf("%s⚠ Sistema apresentou algumas instabilidades%s\n", colorYellow, colorReset)
	} else {
		fmt.Printf("%s✗ Sistema apresentou problemas significativos%s\n", colorRed, colorReset)
	}
}
