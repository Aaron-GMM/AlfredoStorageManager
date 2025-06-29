# AlfredoStorageManager

## Visão Geral

**AlfredoStorageManager** é um gerenciador de arquivos web-based, construído com Go no backend e HTML/CSS/JavaScript no frontend.
 uma interface limpa e intuitiva para interagir com o sistema de arquivos do meu servidor, permitindo listar, navegar por diretórios e criar novas pastas.
 Projetado para ser leve e eficiente, pois meu servidor é modesto em hardware muito simples, precisava de um solução simples leves e facil de ultilizar.
 deploy feito  no proprio servidor, como são arquivos pessoais somente eu tenho acesso e dispositivos que eu escolho

## Inspiração

O nome "Alfredo" é uma homenagem ao icônico mordomo do Batman, Alfred Pennyworth,
 refletindo a ideia de um assistente leal e capaz que gerencia os recursos de forma eficiente e discreta. 

## Funcionalidades

* **Listagem de Arquivos e Diretórios:** Visualize facilmente o conteúdo das pastas.
* **Navegação Intuitiva:** Clique em diretórios para explorá-los e utilize o botão "Voltar" para subir na hierarquia.
* **Criação de Pastas:** Crie novas pastas de forma simples através da interface web.
* **Configuração de Caminho Base:** O diretório raiz do gerenciamento de arquivos é configurável via variáveis de ambiente, garantindo flexibilidade e segurança.



## Estrutura do Projeto
├── go_server/
│   ├── api/
│   │   ├── .env              # Arquivo de variáveis de ambiente (local)
│   │   ├── main.go           # Código do servidor Go (backend)
|   |   ├── .gitignore        # Arquivo para ignorar arquivos no Git 
│   │   ├──  go.mod           # Módulo Go e dependências
        └── README.md         # Este arquivo!
       
│   ├── Frontend/
│   │   ├── index.html        # Interface do usuário (frontend)
│   │   ├── style.css         # Estilos CSS
│   │   └── script.js         # Lógica JavaScript


