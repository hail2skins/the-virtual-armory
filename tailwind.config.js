/** @type {import('tailwindcss').Config} */
module.exports = {
    content: [
               "./cmd/web/**/*.html", "./cmd/web/**/*.templ",
    ],
    theme: {
        extend: {
            colors: {
                gunmetal: {
                    50: '#f5f6f7',
                    100: '#e5e7e9',
                    200: '#cbd0d5',
                    300: '#a5adb7',
                    400: '#768494',
                    500: '#5a6978',
                    600: '#475563',
                    700: '#3a4550',
                    800: '#333c45',
                    900: '#2c333b',
                    950: '#1a1f24',
                },
                rust: {
                    500: '#8B4513',
                    600: '#723A10',
                    700: '#5A2E0D',
                },
                brass: {
                    300: '#D6C8A4',
                    400: '#C5B076',
                    500: '#B5A55D',
                },
            },
        },
    },
    plugins: [],
}

