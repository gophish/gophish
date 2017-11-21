'use strict';

(function () {
  var TO_RADIANS = Math.PI / 180;

  CKEDITOR.plugins.add('imagerotate', {
    lang: 'en,et',
    hidpi: true,
    icons: 'rotate-left.png,rotate-right.png', // is this line necessary?
    init: function (editor) {

      editor.addCommand('rotateLeft', {
        exec: function (editor) {
          rotateSelectedImageByAngle(editor, -90);
        }
      });

      editor.addCommand('rotateRight', {
        exec: function (editor) {
          rotateSelectedImageByAngle(editor, 90);
        }
      });

      var translations = editor.lang.imagerotate;

      if (!translations) {
        translations = {
          rotateRight: "Rotate Clockwise",
          rotateLeft: "Rotate Counter-clockwise"
        }
      }

      if (editor.contextMenu) {
        editor.addMenuItems({
          rotateRight: {
            label: translations.rotateRight,
            icon: this.path + 'icons/rotate-right.png',
            command: 'rotateRight',
            group: 'image',
            order: 1
          },
          rotateLeft: {
            label: translations.rotateLeft,
            icon: this.path + 'icons/rotate-left.png',
            command: 'rotateLeft',
            group: 'image',
            order: 2
          }
        });

        editor.contextMenu.addListener(function (element, selection) {
          var imageElement = element.getAscendant('img', true);
          if (imageElement) {
            return {
              rotateLeft: CKEDITOR.TRISTATE_OFF,
              rotateRight: CKEDITOR.TRISTATE_OFF
            };
          }
        });
      }

    }
  });


  function rotateSelectedImageByAngle(editor, angle) {
    var selection = editor.getSelection();
    var element = selection.getStartElement();
    var imageElement = element.getAscendant('img', true);
    if (!imageElement) {
      editor.showNotification("no image element?", "warning");
      return;
    }
    var domImageElement = imageElement.$;
    if (!domImageElement) {
      editor.showNotification("no DOM image element?", "warning");
      return;
    }

    if ('crossOrigin' in domImageElement) {
      // this will not work if image respond headers will not have Access-Control-Allow-Origin: *
      domImageElement.setAttribute("crossOrigin", "anonymous");
    }

    try {
      rotateByAngle(domImageElement, angle);
    } catch (err) {
      if (err.code === 18) {
        editor.showNotification("Image is from other domain and can't be rotated", "warning");
      }
    }
  }

  function rotateByAngle(imageElement, angle) {
    var canvas = createRotatedImageCanvas(imageElement, angle);
    //document.body.appendChild(canvas); // dont use this, it's for debugging only
    _putCanvasToImage(canvas, imageElement);
  }

  function _putCanvasToImage(canvas, image) {
    var dataURL = canvas.toDataURL();
    image.src = dataURL;
    image.setAttribute("data-cke-saved-src", dataURL);

    image.style.width = canvas.width + "px";
    image.style.height = canvas.height + "px";

    // TODO: should I support width and height attributes explicitly if inline style is disallowed?
    // see http://docs.ckeditor.com/#!/guide/dev_acf-section-example%3A-disallow-inline-styles-and-use-attributes-instead
    //image.width = canvas.width;
    //image.height = canvas.height;
  }

  function createRotatedImageCanvas(image, angle) {
    angle = angle < 0 ? 360 + angle : angle;

    var canvasWidth = image.width;
    var canvasHeight = image.height;
    if (angle == 90 || angle == 270) {
      canvasWidth = image.height;
      canvasHeight = image.width;
    }

    var x = canvasWidth / 2;
    var y = canvasHeight / 2;
    var width = image.width;
    var height = image.height;

    var canvas = _createCanvas(canvasWidth, canvasHeight);
    var context = canvas.getContext('2d');
    context.translate(x, y);
    context.rotate(angle * TO_RADIANS);
    context.drawImage(image, -width / 2, -height / 2, width, height);
    return canvas;
  }

  function _createCanvas(width, height) {
    var canvas = document.createElement('canvas');
    canvas.width = width;
    canvas.height = height;
    return canvas;
  }

})();