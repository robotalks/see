(function (exports) {
    'use strict';

    vis.defineCanvasObject('camera', {
        paint: function (rc, canvas) {
            var angle = this.properties.angle;
            if (angle == null) {
                angle = 0;
            }
            var x0 = rc.w >> 1;
            var y0 = rc.h >> 1;
            var r = Math.min(Math.min(rc.w / 2, rc.h / 2), Math.sqrt(rc.w * rc.w + rc.h * rc.h) / 2);
            var t = (45 + angle) * Math.PI / 180;
            var x1 = Math.ceil(r * Math.cos(t));
            var y1 = Math.ceil(r * Math.sin(t));
            var t = (-45 + angle) * Math.PI / 180;
            var x2 = Math.ceil(r * Math.cos(t));
            var y2 = Math.ceil(r * Math.sin(t));

            var ctx = canvas.getContext('2d');
            ctx.clearRect(0, 0, rc.w, rc.h);
            ctx.beginPath();
            ctx.moveTo(x0, y0);
            ctx.lineTo(x0 + x1, y0 - y1);
            ctx.lineTo(x0 + x2, y0 - y2);
            ctx.closePath();
            ctx.stroke();
            ctx.setLineDash([1, 1]);
            ctx.beginPath();
            ctx.moveTo(0, y0);
            ctx.lineTo(rc.w, y0);
            ctx.moveTo(x0, 0);
            ctx.lineTo(x0, rc.h);
            ctx.stroke();
        }
    });
})(window);
