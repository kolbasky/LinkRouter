const icons = import.meta.glob('./assets/icons/*.svg', { 
  query: { url: true },
  eager: true,
  import: 'default'
})

export function getIconUrl(iconName) {
  const path = `./assets/icons/${iconName}.svg`
  return icons[path] || ''
}
