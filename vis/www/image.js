(function(exports) {
    'use strict';

    vis.defineObject('image', {
        createContent: function () {
            this._image = document.createElement('img');
            this.update(this.properties);
            return this._image;
        },

        update: function (props) {
            if (this._image) {
                this._image.setAttribute("src", props.src);
                this.properties = props;
                return true;
            }
            return false;
        }
    });
})(window);
