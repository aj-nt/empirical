<?xml version="1.0" encoding="UTF-8"?>
<xsl:stylesheet xmlns:xsl="http://www.w3.org/1999/XSL/Transform"
                xmlns:math="http://exslt.org/math"
                xmlns:exslt="http://exslt.org/common"
                xmlns:koine="urn:empirical:koine"
                version="1.0"
                exclude-result-prefixes="math exslt">

  <!-- ── Parameters ────────────────────────────────────────────────────── -->
  <xsl:param name="orb" select="5.0"/>

  <!-- ── Classical planets (Koiné uses 7) ─────────────────────────────── -->
  <!-- Comma-delimited string for XSLT 1.0 contains() checks -->
  <xsl:variable name="classicalPlanets" select="',Sun,Moon,Mercury,Venus,Mars,Jupiter,Saturn,'"/>

  <!-- ── Ptolemaic aspects ────────────────────────────────────────────── -->
  <xsl:variable name="aspects">
    <aspect angle="0" name="conjunction"/>
    <aspect angle="60" name="sextile"/>
    <aspect angle="90" name="square"/>
    <aspect angle="120" name="trine"/>
    <aspect angle="180" name="opposition"/>
  </xsl:variable>

  <!-- ── Sign names (comma-delimited for XSLT 1.0) ────────────────────── -->
  <xsl:variable name="signs" select="'Aries,Taurus,Gemini,Cancer,Leo,Virgo,Libra,Scorpio,Sagittarius,Capricorn,Aquarius,Pisces'"/>

  <!-- ── Domicile rulers (planet → sign index, 0-based) ───────────────── -->
  <xsl:variable name="domicileSun" select="4"/>     <!-- Leo -->
  <xsl:variable name="domicileMoon" select="3"/>    <!-- Cancer -->
  <xsl:variable name="domicileMercury" select="2"/> <!-- Gemini (also Virgo=5) -->
  <xsl:variable name="domicileVenus" select="1"/>   <!-- Taurus (also Libra=6) -->
  <xsl:variable name="domicileMars" select="0"/>    <!-- Aries (also Scorpio=7) -->
  <xsl:variable name="domicileJupiter" select="8"/> <!-- Sagittarius (also Pisces=11) -->
  <xsl:variable name="domicileSaturn" select="9"/>  <!-- Capricorn (also Aquarius=10) -->

  <!-- ── Exaltation degrees (planet → sign index + degree) ────────────── -->
  <xsl:variable name="exaltationSun" select="0"/>     <!-- Aries 19° -->
  <xsl:variable name="exaltationSunDeg" select="19"/>
  <xsl:variable name="exaltationMoon" select="1"/>    <!-- Taurus 3° -->
  <xsl:variable name="exaltationMoonDeg" select="3"/>
  <xsl:variable name="exaltationMercury" select="5"/> <!-- Virgo 15° -->
  <xsl:variable name="exaltationMercuryDeg" select="15"/>
  <xsl:variable name="exaltationVenus" select="11"/>   <!-- Pisces 27° -->
  <xsl:variable name="exaltationVenusDeg" select="27"/>
  <xsl:variable name="exaltationMars" select="9"/>     <!-- Capricorn 28° -->
  <xsl:variable name="exaltationMarsDeg" select="28"/>
  <xsl:variable name="exaltationJupiter" select="3"/>  <!-- Cancer 15° -->
  <xsl:variable name="exaltationJupiterDeg" select="15"/>
  <xsl:variable name="exaltationSaturn" select="6"/>    <!-- Libra 21° -->
  <xsl:variable name="exaltationSaturnDeg" select="21"/>

  <!-- ── Triplicity rulers (fire/earth/air/water, day/night/participating) ── -->
  <xsl:variable name="triplicityFireDay" select="'Sun'"/>
  <xsl:variable name="triplicityFireNight" select="'Jupiter'"/>
  <xsl:variable name="triplicityFirePart" select="'Saturn'"/>
  <xsl:variable name="triplicityEarthDay" select="'Venus'"/>
  <xsl:variable name="triplicityEarthNight" select="'Moon'"/>
  <xsl:variable name="triplicityEarthPart" select="'Mars'"/>
  <xsl:variable name="triplicityAirDay" select="'Saturn'"/>
  <xsl:variable name="triplicityAirNight" select="'Mercury'"/>
  <xsl:variable name="triplicityAirPart" select="'Jupiter'"/>
  <xsl:variable name="triplicityWaterDay" select="'Venus'"/>
  <xsl:variable name="triplicityWaterNight" select="'Mars'"/>
  <xsl:variable name="triplicityWaterPart" select="'Moon'"/>

  <!-- ── Root template ────────────────────────────────────────────────── -->
  <xsl:template match="/BaseChart">
    <koine:Chart xmlns:koine="urn:empirical:koine">
      <koine:Name><xsl:value-of select="Identity/Name"/></koine:Name>

      <!-- Determine sect -->
      <xsl:variable name="sunLon" select="Positions/Planet[@name='Sun']/Tropical/Lon"/>
      <xsl:variable name="ascLon" select="Angles/ASC"/>
      <xsl:variable name="sunAscDiff" select="($sunLon - $ascLon + 360) mod 360"/>
      <xsl:variable name="isDay" select="$sunAscDiff &lt; 180"/>

      <koine:Sect>
        <xsl:choose>
          <xsl:when test="$isDay">day</xsl:when>
          <xsl:otherwise>night</xsl:otherwise>
        </xsl:choose>
      </koine:Sect>

      <!-- Planet signs -->
      <koine:PlanetSigns>
        <xsl:for-each select="Positions/Planet[contains($classicalPlanets, concat(',', @name, ','))]">
          <xsl:sort select="@id" data-type="number"/>
          <koine:PlanetSign>
            <koine:Planet><xsl:value-of select="@name"/></koine:Planet>
            <koine:Sign>
              <xsl:call-template name="lonToSign">
                <xsl:with-param name="lon" select="Tropical/Lon"/>
              </xsl:call-template>
            </koine:Sign>
            <koine:Degree>
              <xsl:value-of select="format-number(Tropical/Lon mod 30, '0.00')"/>
            </koine:Degree>
            <koine:Retrograde>
              <xsl:value-of select="Tropical/Speed &lt; 0"/>
            </koine:Retrograde>
          </koine:PlanetSign>
        </xsl:for-each>
      </koine:PlanetSigns>

      <!-- Whole-sign houses -->
      <koine:PlanetHouses>
        <xsl:for-each select="Positions/Planet[contains($classicalPlanets, concat(',', @name, ','))]">
          <xsl:sort select="@id" data-type="number"/>
          <koine:PlanetHouse>
            <koine:Planet><xsl:value-of select="@name"/></koine:Planet>
            <koine:House>
              <xsl:call-template name="lonToHouse">
                <xsl:with-param name="lon" select="Tropical/Lon"/>
                <xsl:with-param name="asc" select="$ascLon"/>
              </xsl:call-template>
            </koine:House>
          </koine:PlanetHouse>
        </xsl:for-each>
      </koine:PlanetHouses>

      <!-- Aspects (Ptolemaic, classical planets only) -->
      <koine:Aspects>
        <xsl:call-template name="computeAspects">
          <xsl:with-param name="planets" select="Positions/Planet[contains($classicalPlanets, concat(',', @name, ','))]"/>
        </xsl:call-template>
      </koine:Aspects>

      <!-- Essential dignities -->
      <koine:Dignities>
        <xsl:for-each select="Positions/Planet[contains($classicalPlanets, concat(',', @name, ','))]">
          <xsl:sort select="@id" data-type="number"/>
          <xsl:call-template name="assessDignity">
            <xsl:with-param name="planet" select="."/>
            <xsl:with-param name="isDay" select="$isDay"/>
          </xsl:call-template>
        </xsl:for-each>
      </koine:Dignities>

      <!-- Nodes -->
      <koine:Nodes>
        <koine:NorthNode>
          <koine:Sign>
            <xsl:call-template name="lonToSign">
              <xsl:with-param name="lon" select="Nodes/NorthNode"/>
            </xsl:call-template>
          </koine:Sign>
          <koine:Degree>
            <xsl:value-of select="format-number(Nodes/NorthNode mod 30, '0.00')"/>
          </koine:Degree>
        </koine:NorthNode>
        <koine:SouthNode>
          <koine:Sign>
            <xsl:call-template name="lonToSign">
              <xsl:with-param name="lon" select="Nodes/SouthNode"/>
            </xsl:call-template>
          </koine:Sign>
          <koine:Degree>
            <xsl:value-of select="format-number(Nodes/SouthNode mod 30, '0.00')"/>
          </koine:Degree>
        </koine:SouthNode>
      </koine:Nodes>

    </koine:Chart>
  </xsl:template>

  <!-- ── Helper: longitude to sign name ───────────────────────────────── -->
  <xsl:template name="lonToSign">
    <xsl:param name="lon"/>
    <xsl:variable name="idx" select="floor($lon div 30)"/>
    <xsl:call-template name="nthToken">
      <xsl:with-param name="list" select="'Aries,Taurus,Gemini,Cancer,Leo,Virgo,Libra,Scorpio,Sagittarius,Capricorn,Aquarius,Pisces'"/>
      <xsl:with-param name="n" select="$idx + 1"/>
    </xsl:call-template>
  </xsl:template>

  <!-- ── Helper: nth comma-separated token ────────────────────────────── -->
  <xsl:template name="nthToken">
    <xsl:param name="list"/>
    <xsl:param name="n"/>
    <xsl:choose>
      <xsl:when test="$n = 1">
        <xsl:value-of select="substring-before(concat($list, ','), ',')"/>
      </xsl:when>
      <xsl:otherwise>
        <xsl:call-template name="nthToken">
          <xsl:with-param name="list" select="substring-after($list, ',')"/>
          <xsl:with-param name="n" select="$n - 1"/>
        </xsl:call-template>
      </xsl:otherwise>
    </xsl:choose>
  </xsl:template>

  <!-- ── Helper: longitude to whole-sign house ────────────────────────── -->
  <xsl:template name="lonToHouse">
    <xsl:param name="lon"/>
    <xsl:param name="asc"/>
    <xsl:variable name="house" select="(floor($lon div 30) - floor($asc div 30) + 12) mod 12 + 1"/>
    <xsl:value-of select="$house"/>
  </xsl:template>

  <!-- ── Helper: compute aspects between planet set ───────────────────── -->
  <xsl:template name="computeAspects">
    <xsl:param name="planets"/>
    <xsl:for-each select="$planets">
      <xsl:variable name="p1pos" select="position()"/>
      <xsl:variable name="p1" select="."/>
      <xsl:variable name="p1lon" select="$p1/Tropical/Lon"/>
      <xsl:variable name="p1name" select="$p1/@name"/>
      <xsl:for-each select="$planets[position() > $p1pos]">
        <xsl:variable name="p2lon" select="Tropical/Lon"/>
        <xsl:variable name="p2name" select="@name"/>
        <xsl:variable name="dist" select="($p2lon - $p1lon + 360) mod 360"/>
        <!-- Check each Ptolemaic aspect angle inline -->
        <xsl:call-template name="checkAspect">
          <xsl:with-param name="p1name" select="$p1name"/>
          <xsl:with-param name="p2name" select="$p2name"/>
          <xsl:with-param name="dist" select="$dist"/>
          <xsl:with-param name="angle" select="0"/>
          <xsl:with-param name="aspectName" select="'conjunction'"/>
        </xsl:call-template>
        <xsl:call-template name="checkAspect">
          <xsl:with-param name="p1name" select="$p1name"/>
          <xsl:with-param name="p2name" select="$p2name"/>
          <xsl:with-param name="dist" select="$dist"/>
          <xsl:with-param name="angle" select="60"/>
          <xsl:with-param name="aspectName" select="'sextile'"/>
        </xsl:call-template>
        <xsl:call-template name="checkAspect">
          <xsl:with-param name="p1name" select="$p1name"/>
          <xsl:with-param name="p2name" select="$p2name"/>
          <xsl:with-param name="dist" select="$dist"/>
          <xsl:with-param name="angle" select="90"/>
          <xsl:with-param name="aspectName" select="'square'"/>
        </xsl:call-template>
        <xsl:call-template name="checkAspect">
          <xsl:with-param name="p1name" select="$p1name"/>
          <xsl:with-param name="p2name" select="$p2name"/>
          <xsl:with-param name="dist" select="$dist"/>
          <xsl:with-param name="angle" select="120"/>
          <xsl:with-param name="aspectName" select="'trine'"/>
        </xsl:call-template>
        <xsl:call-template name="checkAspect">
          <xsl:with-param name="p1name" select="$p1name"/>
          <xsl:with-param name="p2name" select="$p2name"/>
          <xsl:with-param name="dist" select="$dist"/>
          <xsl:with-param name="angle" select="180"/>
          <xsl:with-param name="aspectName" select="'opposition'"/>
        </xsl:call-template>
      </xsl:for-each>
    </xsl:for-each>
  </xsl:template>

  <!-- ── Check a single aspect angle ──────────────────────────────────── -->
  <xsl:template name="checkAspect">
    <xsl:param name="p1name"/>
    <xsl:param name="p2name"/>
    <xsl:param name="dist"/>
    <xsl:param name="angle"/>
    <xsl:param name="aspectName"/>
    <xsl:variable name="diff" select="$dist - $angle"/>
    <xsl:variable name="absDiff">
      <xsl:choose>
        <xsl:when test="$diff &lt; 0"><xsl:value-of select="-$diff"/></xsl:when>
        <xsl:otherwise><xsl:value-of select="$diff"/></xsl:otherwise>
      </xsl:choose>
    </xsl:variable>
    <xsl:variable name="oppDiff" select="360 - $dist - $angle"/>
    <xsl:variable name="absOppDiff">
      <xsl:choose>
        <xsl:when test="$oppDiff &lt; 0"><xsl:value-of select="-$oppDiff"/></xsl:when>
        <xsl:otherwise><xsl:value-of select="$oppDiff"/></xsl:otherwise>
      </xsl:choose>
    </xsl:variable>
    <xsl:if test="$absDiff &lt;= $orb or $absOppDiff &lt;= $orb">
      <koine:Aspect>
        <koine:Planet1><xsl:value-of select="$p1name"/></koine:Planet1>
        <koine:Planet2><xsl:value-of select="$p2name"/></koine:Planet2>
        <koine:Type><xsl:value-of select="$aspectName"/></koine:Type>
        <koine:Orb>
          <xsl:choose>
            <xsl:when test="$absDiff &lt;= $absOppDiff">
              <xsl:value-of select="format-number($absDiff, '0.00')"/>
            </xsl:when>
            <xsl:otherwise>
              <xsl:value-of select="format-number($absOppDiff, '0.00')"/>
            </xsl:otherwise>
          </xsl:choose>
        </koine:Orb>
      </koine:Aspect>
    </xsl:if>
  </xsl:template>

  <!-- ── Helper: assess essential dignity ─────────────────────────────── -->
  <xsl:template name="assessDignity">
    <xsl:param name="planet"/>
    <xsl:param name="isDay"/>
    <xsl:variable name="name" select="$planet/@name"/>
    <xsl:variable name="lon" select="$planet/Tropical/Lon"/>
    <xsl:variable name="signIdx" select="floor($lon div 30)"/>
    <xsl:variable name="degree" select="$lon mod 30"/>

    <koine:Dignity>
      <koine:Planet><xsl:value-of select="$name"/></koine:Planet>
      <koine:Sign>
        <xsl:call-template name="nthToken">
          <xsl:with-param name="list" select="'Aries,Taurus,Gemini,Cancer,Leo,Virgo,Libra,Scorpio,Sagittarius,Capricorn,Aquarius,Pisces'"/>
          <xsl:with-param name="n" select="$signIdx + 1"/>
        </xsl:call-template>
      </koine:Sign>

      <!-- Domicile -->
      <koine:Domicile>
        <xsl:call-template name="checkDomicile">
          <xsl:with-param name="planet" select="$name"/>
          <xsl:with-param name="signIdx" select="$signIdx"/>
        </xsl:call-template>
      </koine:Domicile>

      <!-- Exaltation -->
      <koine:Exaltation>
        <xsl:call-template name="checkExaltation">
          <xsl:with-param name="planet" select="$name"/>
          <xsl:with-param name="signIdx" select="$signIdx"/>
          <xsl:with-param name="degree" select="$degree"/>
        </xsl:call-template>
      </koine:Exaltation>

      <!-- Triplicity -->
      <koine:Triplicity>
        <xsl:call-template name="checkTriplicity">
          <xsl:with-param name="planet" select="$name"/>
          <xsl:with-param name="signIdx" select="$signIdx"/>
          <xsl:with-param name="isDay" select="$isDay"/>
        </xsl:call-template>
      </koine:Triplicity>

      <!-- State -->
      <koine:State>
        <xsl:call-template name="determineState">
          <xsl:with-param name="planet" select="$name"/>
          <xsl:with-param name="signIdx" select="$signIdx"/>
          <xsl:with-param name="degree" select="$degree"/>
          <xsl:with-param name="isDay" select="$isDay"/>
        </xsl:call-template>
      </koine:State>
    </koine:Dignity>
  </xsl:template>

  <!-- ── Domicile check ───────────────────────────────────────────────── -->
  <xsl:template name="checkDomicile">
    <xsl:param name="planet"/>
    <xsl:param name="signIdx"/>
    <xsl:choose>
      <xsl:when test="$planet = 'Sun' and $signIdx = 4">true</xsl:when>
      <xsl:when test="$planet = 'Moon' and $signIdx = 3">true</xsl:when>
      <xsl:when test="$planet = 'Mercury' and ($signIdx = 2 or $signIdx = 5)">true</xsl:when>
      <xsl:when test="$planet = 'Venus' and ($signIdx = 1 or $signIdx = 6)">true</xsl:when>
      <xsl:when test="$planet = 'Mars' and ($signIdx = 0 or $signIdx = 7)">true</xsl:when>
      <xsl:when test="$planet = 'Jupiter' and ($signIdx = 8 or $signIdx = 11)">true</xsl:when>
      <xsl:when test="$planet = 'Saturn' and ($signIdx = 9 or $signIdx = 10)">true</xsl:when>
      <xsl:otherwise>false</xsl:otherwise>
    </xsl:choose>
  </xsl:template>

  <!-- ── Exaltation check ─────────────────────────────────────────────── -->
  <xsl:template name="checkExaltation">
    <xsl:param name="planet"/>
    <xsl:param name="signIdx"/>
    <xsl:param name="degree"/>
    <xsl:choose>
      <xsl:when test="$planet = 'Sun' and $signIdx = 0">true</xsl:when>
      <xsl:when test="$planet = 'Moon' and $signIdx = 1">true</xsl:when>
      <xsl:when test="$planet = 'Mercury' and $signIdx = 5">true</xsl:when>
      <xsl:when test="$planet = 'Venus' and $signIdx = 11">true</xsl:when>
      <xsl:when test="$planet = 'Mars' and $signIdx = 9">true</xsl:when>
      <xsl:when test="$planet = 'Jupiter' and $signIdx = 3">true</xsl:when>
      <xsl:when test="$planet = 'Saturn' and $signIdx = 6">true</xsl:when>
      <xsl:otherwise>false</xsl:otherwise>
    </xsl:choose>
  </xsl:template>

  <!-- ── Triplicity check ─────────────────────────────────────────────── -->
  <xsl:template name="checkTriplicity">
    <xsl:param name="planet"/>
    <xsl:param name="signIdx"/>
    <xsl:param name="isDay"/>
    <xsl:variable name="element">
      <xsl:choose>
        <xsl:when test="$signIdx = 0 or $signIdx = 4 or $signIdx = 8">fire</xsl:when>
        <xsl:when test="$signIdx = 1 or $signIdx = 5 or $signIdx = 9">earth</xsl:when>
        <xsl:when test="$signIdx = 2 or $signIdx = 6 or $signIdx = 10">air</xsl:when>
        <xsl:when test="$signIdx = 3 or $signIdx = 7 or $signIdx = 11">water</xsl:when>
      </xsl:choose>
    </xsl:variable>
    <xsl:choose>
      <xsl:when test="$element = 'fire' and $isDay and $planet = 'Sun'">true</xsl:when>
      <xsl:when test="$element = 'fire' and not($isDay) and $planet = 'Jupiter'">true</xsl:when>
      <xsl:when test="$element = 'earth' and $isDay and $planet = 'Venus'">true</xsl:when>
      <xsl:when test="$element = 'earth' and not($isDay) and $planet = 'Moon'">true</xsl:when>
      <xsl:when test="$element = 'air' and $isDay and $planet = 'Saturn'">true</xsl:when>
      <xsl:when test="$element = 'air' and not($isDay) and $planet = 'Mercury'">true</xsl:when>
      <xsl:when test="$element = 'water' and $isDay and $planet = 'Venus'">true</xsl:when>
      <xsl:when test="$element = 'water' and not($isDay) and $planet = 'Mars'">true</xsl:when>
      <xsl:otherwise>false</xsl:otherwise>
    </xsl:choose>
  </xsl:template>

  <!-- ── State determination ──────────────────────────────────────────── -->
  <xsl:template name="determineState">
    <xsl:param name="planet"/>
    <xsl:param name="signIdx"/>
    <xsl:param name="degree"/>
    <xsl:param name="isDay"/>
    <xsl:variable name="domicile">
      <xsl:call-template name="checkDomicile">
        <xsl:with-param name="planet" select="$planet"/>
        <xsl:with-param name="signIdx" select="$signIdx"/>
      </xsl:call-template>
    </xsl:variable>
    <xsl:variable name="exaltation">
      <xsl:call-template name="checkExaltation">
        <xsl:with-param name="planet" select="$planet"/>
        <xsl:with-param name="signIdx" select="$signIdx"/>
        <xsl:with-param name="degree" select="$degree"/>
      </xsl:call-template>
    </xsl:variable>
    <xsl:variable name="triplicity">
      <xsl:call-template name="checkTriplicity">
        <xsl:with-param name="planet" select="$planet"/>
        <xsl:with-param name="signIdx" select="$signIdx"/>
        <xsl:with-param name="isDay" select="$isDay"/>
      </xsl:call-template>
    </xsl:variable>
    <xsl:choose>
      <xsl:when test="$domicile = 'true'">domicile</xsl:when>
      <xsl:when test="$exaltation = 'true'">exaltation</xsl:when>
      <xsl:when test="$triplicity = 'true'">triplicity</xsl:when>
      <xsl:otherwise>peregrine</xsl:otherwise>
    </xsl:choose>
  </xsl:template>

</xsl:stylesheet>
