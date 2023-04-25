window.onload = () => {
  const delay = 5000;
  let id;
  checkbox = document.querySelector("#reload");

  function reload() {
    window.location.reload();
  }

  checkbox.onchange = function () {
    if (this.checked && id === undefined) {
      console.log("set interval");
      id = setInterval(reload, delay);
      sessionStorage.setItem("reload", "true");
    }
    if (!this.checked && id !== undefined) {
      console.log("clear interval");
      sessionStorage.removeItem("reload");
      clearInterval(id);
    }
  };

  if (sessionStorage.getItem("reload")) {
    id = setInterval(reload, delay);
    checkbox.checked = true;
  }
};
