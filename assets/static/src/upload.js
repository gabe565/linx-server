export const initUpload = async () => {
  const dropzone = await import("dropzone");
  const myDropzone = new dropzone.default("#dropzone", {
    init: () => {
      const dzone = document.getElementById("dzone");
      dzone.style.display = "block";
    },
    addedfile: (file) => {
      if (!myDropzone.options.autoProcessQueue) {
        const xhr = new XMLHttpRequest();
        xhr.onload = () => {
          if (xhr.readyState !== XMLHttpRequest.DONE) {
            return;
          }
          if (xhr.status < 400) {
            myDropzone.processQueue();
            myDropzone.options.autoProcessQueue = true;
          } else {
            myDropzone.cancelUpload(file);
          }
        };
        xhr.open("HEAD", "auth", true);
        xhr.send();
      }
      const upload = document.createElement("div");
      upload.className = "upload";

      const fileLabel = document.createElement("span");
      fileLabel.innerHTML = file.name;
      file.fileLabel = fileLabel;
      upload.appendChild(fileLabel);

      const fileActions = document.createElement("div");
      fileActions.className = "right";
      file.fileActions = fileActions;
      upload.appendChild(fileActions);

      const cancelAction = document.createElement("span");
      cancelAction.className = "cancel";
      cancelAction.innerHTML = "Cancel";
      cancelAction.addEventListener("click", () => {
        myDropzone.removeFile(file);
      });
      file.cancelActionElement = cancelAction;
      fileActions.appendChild(cancelAction);

      const progress = document.createElement("span");
      file.progressElement = progress;
      fileActions.appendChild(progress);

      file.uploadElement = upload;

      document.getElementById("uploads").appendChild(upload);
    },
    uploadprogress: (file, p) => {
      p = parseInt(p);
      file.progressElement.innerHTML = p + "%";
      file.uploadElement.setAttribute(
        "style",
        "background-image: -webkit-linear-gradient(left, #F2F4F7 " +
          p +
          "%, #E2E2E2 " +
          p +
          "%); background-image: -moz-linear-gradient(left, #F2F4F7 " +
          p +
          "%, #E2E2E2 " +
          p +
          "%); background-image: -ms-linear-gradient(left, #F2F4F7 " +
          p +
          "%, #E2E2E2 " +
          p +
          "%); background-image: -o-linear-gradient(left, #F2F4F7 " +
          p +
          "%, #E2E2E2 " +
          p +
          "%); background-image: linear-gradient(left, #F2F4F7 " +
          p +
          "%, #E2E2E2 " +
          p +
          "%)",
      );
    },
    sending: (file, xhr, formData) => {
      const randomize = document.getElementById("randomize");
      if (randomize != null) {
        formData.append("randomize", randomize.checked);
      }
      formData.append("expires", document.getElementById("expires").value);
    },
    success: (file, resp) => {
      file.fileActions.removeChild(file.progressElement);

      const fileLabelLink = document.createElement("a");
      fileLabelLink.href = resp.url;
      fileLabelLink.target = "_blank";
      fileLabelLink.innerHTML = resp.url;
      file.fileLabel.innerHTML = "";
      file.fileLabelLink = fileLabelLink;
      file.fileLabel.appendChild(fileLabelLink);

      const deleteAction = document.createElement("span");
      deleteAction.innerHTML = "Delete";
      deleteAction.className = "cancel";
      deleteAction.addEventListener("click", () => {
        const xhr = new XMLHttpRequest();
        xhr.open("DELETE", resp.url, true);
        xhr.setRequestHeader("Linx-Delete-Key", resp.delete_key);
        xhr.onreadystatechange = () => {
          if (xhr.readyState === 4 && xhr.status === 200) {
            const text = document.createTextNode("Deleted ");
            file.fileLabel.insertBefore(text, file.fileLabelLink);
            file.fileLabel.className = "deleted";
            file.fileActions.removeChild(file.cancelActionElement);
          }
        };
        xhr.send();
      });
      file.fileActions.removeChild(file.cancelActionElement);
      file.cancelActionElement = deleteAction;
      file.fileActions.appendChild(deleteAction);
    },
    canceled: (file) => {
      myDropzone.options.error(file);
    },
    error: (file, resp) => {
      file.fileActions.removeChild(file.cancelActionElement);
      file.fileActions.removeChild(file.progressElement);

      if (file.status === "canceled") {
        file.fileLabel.innerHTML = file.name + ": Canceled ";
      } else {
        if (resp.error) {
          file.fileLabel.innerHTML = file.name + ": " + resp.error;
        } else if (resp.includes("<html")) {
          file.fileLabel.innerHTML = file.name + ": Server Error";
        } else {
          file.fileLabel.innerHTML = file.name + ": " + resp;
        }
      }
      file.fileLabel.className = "error";
    },

    autoProcessQueue: document.getElementById("dropzone").getAttribute("data-auth") !== "basic",
    maxFilesize: Math.round(
      parseInt(document.getElementById("dropzone").getAttribute("data-maxsize"), 10) / 1024 / 1024,
    ),
    previewsContainer: "#uploads",
    parallelUploads: 5,
    headers: { Accept: "application/json" },
    dictDefaultMessage: "Click or Drop file(s) or Paste image",
    dictFallbackMessage: "",
  });

  document.addEventListener("paste", (e) => {
    const items = (e.clipboardData || e.originalEvent.clipboardData).items;
    for (let index in items) {
      const item = items[index];
      if (item.kind === "file") {
        myDropzone.addFile(item.getAsFile());
      }
    }
  });

  document.getElementById("access_key_checkbox").addEventListener("change", (e) => {
    const input = document.getElementById("access_key_input");
    const text = document.getElementById("access_key_text");
    if (e.target.checked) {
      input.style.display = "inline-block";
      text.style.display = "none";
    } else {
      input.value = "";
      input.style.display = "none";
      text.style.display = "inline-block";
    }
  });
};
