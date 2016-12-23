(function(exports) {
    'use strict';

    vis.defineObject('image', {
        createContent: function () {
            this._image = document.createElement('img');
            this._update();
            return this._image;
        },

        update: function (props) {
            this.properties = props;
            if (this._image) {
                this._update();
                return true;
            }
            return false;
        },

        destroy: function () {
            this._stopInterval();
        },

        _update: function () {
            this._stopInterval();
            var src = this.properties.src;
            if (typeof(src) != 'string') {
                src = "";
            }
            src = src.replace('TIMESTAMP', Date.now());
            this._image.setAttribute('src', src);
            if (this.properties.interval != null) {
                this._timer = setTimeout(this._update.bind(this),
                    this.properties.interval);
            }
        },

        _stopInterval: function () {
            if (this._timer) {
                clearTimeout(this._timer);
                delete this._timer;
            }
        }
    });
})(window);
