// echarts 按需注册：只打包项目实际用到的图表和组件，避免 `import * as echarts from "echarts"` 引入全量。
// 新图表用到这里没注册的 series 类型或组件（如 bar、dataZoom）时，要在下面补上，否则不会渲染。
import * as echarts from "echarts/core";
import { LineChart, PieChart } from "echarts/charts";
import {
  GraphicComponent,
  GridComponent,
  LegendComponent,
  MarkAreaComponent,
  TitleComponent,
  TooltipComponent,
} from "echarts/components";
import { CanvasRenderer } from "echarts/renderers";

echarts.use([
  LineChart,
  PieChart,
  GraphicComponent,
  GridComponent,
  LegendComponent,
  MarkAreaComponent,
  TitleComponent,
  TooltipComponent,
  CanvasRenderer,
]);

export default echarts;
