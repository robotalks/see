(function (exports) {
    'use strict';

    vis.defineCanvasObject('chart', {
        destroy: function () {
            if (this._chart) {
                this._chart.destroy();
            }
        },

        update: function (props) {
            this.properties = props;
            if (this._chart) {
                this._chart.config = this._buildChart();
                this._chart.update();
                return true;
            }
            return false;
        },
    
        paint: function (rc, canvas) {
            var conf = this._buildChart();
            if (this._chart) {
                this._chart.config = conf;
                this._chart.update({ duration: 0 });
            } else {
                this._chart = new Chart(canvas, conf);
            }
        },

        _buildChart: function () {
            var conf = {
                type: this.properties.chart || 'line',
                data: {
                    labels: this.properties.labels,
                },
                options: this.properties.options || {},
            };
            if (Array.isArray(this.properties.datasets)) {
                conf.data.datasets = this.properties.datasets.map((ds) => {
                    return {
                        label: this.properties.title,
                        data: this.properties.labels.map((label) => ds[label]),
                        backgroundColor: this.properties.colors,
                    };
                });
            } else {
                conf.data.datasets = [{
                    label: this.properties.title,
                    data: this.properties.labels.map((label) => this.properties.data[label]),
                    backgroundColor: this.properties.colors,
                }];
            }
            return conf;
        }
    });
})(window);
