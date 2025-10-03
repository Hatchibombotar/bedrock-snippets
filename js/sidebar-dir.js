function toggleDir(id) {
  const directoryContent = document.querySelector(`[data-directory-content="${id}"]`)
  const directoryHeader = document.querySelector(`[data-directory-header="${id}"]`)
  
  directoryContent.classList.toggle("hidden")
  directoryHeader.classList.toggle("directory-header-closed")
}
