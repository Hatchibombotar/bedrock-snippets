
const sidebar = document.querySelector("#sidebar")
const sidebarToggle = document.querySelector("#sidebar-toggle")
const sidebarToggleIcon = document.querySelector("#sidebar-toggle-icon")
const mainContent = document.querySelector("#main-content")


let showSidebar = true
sidebarToggle.onclick = () => {
    showSidebar = !showSidebar

    if (showSidebar) {
        sidebar.classList.remove("hidden")
        mainContent.classList.remove("rounded-r")
        sidebarToggleIcon.src = ROOT_DIRECTORY + "/OcSidebarcollapse2.svg"
    } else {
        sidebar.classList.add("hidden")
        mainContent.classList.add("rounded-r")
        sidebarToggleIcon.src = ROOT_DIRECTORY + "/OcSidebarexpand2.svg"
    }
}
