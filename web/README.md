# Argos Panoptes - Web Interface

Interface web para o sistema de monitoramento Argos Panoptes.

## Descrição

Este é o frontend da plataforma de monitoramento DevOps Argos Panoptes, oferecendo uma interface moderna e responsiva para visualização de métricas, status de serviços e alertas em tempo real.

## Tecnologias

- **Vite** - Build tool e dev server
- **TypeScript** - Tipagem estática
- **React** - Framework UI
- **shadcn-ui** - Componentes UI
- **Tailwind CSS** - Estilização
- **React Router** - Roteamento
- **TanStack Query** - Gerenciamento de estado assíncrono
- **Recharts** - Gráficos e visualizações

## Desenvolvimento

### Pré-requisitos

- Node.js 18+ (recomendado: instalar via [nvm](https://github.com/nvm-sh/nvm#installing-and-updating))
- npm ou yarn

### Instalação

```sh
# Instalar dependências
npm install

# Iniciar servidor de desenvolvimento
npm run dev

# Build para produção
npm run build

# Visualizar build de produção
npm run preview
```

## Estrutura do Projeto

```
src/
├── components/     # Componentes reutilizáveis
├── pages/         # Páginas da aplicação
├── hooks/         # Custom hooks
├── lib/           # Utilitários e helpers
└── main.tsx       # Entry point
```

## Funcionalidades

- Dashboard com visão geral do sistema
- Monitoramento de Web Server (Apache, Nginx)
- Monitoramento de Database (PostgreSQL, MySQL)
- Monitoramento de DNS
- Monitoramento de SMTP
- Alertas e notificações em tempo real
- Interface responsiva e moderna
