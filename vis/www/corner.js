(function (exports) {
    'use strict';

    vis.defineCanvasObject('corner', {
        paint: function (viewRc, canvas) {
            var w = viewRc.w, h = viewRc.h;
            var ctx = canvas.getContext('2d');
            ctx.clearRect(0, 0, w, h);
            ctx.strokeStyle = $(canvas).css('color');
            ctx.lineWidth = 1;
            ctx.beginPath();
            switch (this.properties.loc) {
            case 'lt':
                ctx.moveTo(0, h);
                ctx.lineTo(0, 0);
                ctx.lineTo(w, 0);
                break;
            case 'rt':
                ctx.moveTo(0, 0);
                ctx.lineTo(w - 1, 0);
                ctx.lineTo(w - 1, h);
                break;
            case 'lb':
                ctx.moveTo(0, 0);
                ctx.lineTo(0, h - 1);
                ctx.lineTo(w, h - 1);
            case 'rb':
                ctx.moveTo(0, h - 1);
                ctx.lineTo(w - 1, h - 1);
                ctx.lineTo(w - 1, 0);
            }
            ctx.stroke();
        }
    });
})(window);
