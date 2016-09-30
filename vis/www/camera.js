(function (exports) {
    'use strict';

    vis.defineCanvasObject('camera', {
        paint: function (rc, canvas) {
            var angle = this.properties.angle;
            if (angle == null) {
                angle = 0;
            }

            var w = rc.w >> 1, h = rc.h >> 1;
            var r = Math.min(w, h);
            var dw = r * Math.cos(30 * Math.PI / 180);
            var dh = r * Math.sin(30 * Math.PI / 180);

            var ctx = canvas.getContext('2d');
            ctx.clearRect(0, 0, rc.w, rc.h);

            ctx.translate(w, h);
            ctx.beginPath();
            ctx.moveTo(0, -h);
            ctx.lineTo(0, h);
            ctx.moveTo(-w, 0);
            ctx.lineTo(w, 0);
            ctx.setLineDash([1, 1]);
            ctx.stroke();

            if (angle != 0) {
                ctx.rotate(-angle * Math.PI / 180);
            }
            ctx.beginPath();
            ctx.moveTo(dw, -dh);
            ctx.lineTo(0, 0);
            ctx.lineTo(dw, dh);
            ctx.closePath();
            ctx.setLineDash([1, 0]);
            ctx.stroke();
        }
    });
})(window);
